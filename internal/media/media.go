// Package media caches article images and serves them to the WebView through
// a local MyGO protocol.
package media

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

// Scheme is the local protocol used by rewritten article images.
const Scheme = "ressu-media"

// MaxBytes bounds a single downloaded image.
const MaxBytes = 8 << 20

// MaxCacheBytes bounds the disk space used by downloaded image bytes.
const MaxCacheBytes = 256 << 20

// MaxEntries bounds source mappings created for one local cache.
const MaxEntries = 10000

const (
	freshTTL    = 7 * 24 * time.Hour
	failureTTL  = time.Hour
	indexPath   = "index"
	dataPath    = "data"
	dataSuffix  = ".bin"
	maxRedirect = 4
	maxFailures = 10000
)

var loggedURLPattern = regexp.MustCompile(`(?i)([a-z][a-z0-9+.-]*://)[^\s"'<>]+`)

// Entry describes one cached image. A nil size marks a prepared, but not yet
// downloaded resource.
type Entry struct {
	Key          string    `json:"key"`
	SourceURL    string    `json:"sourceUrl"`
	ContentType  string    `json:"contentType"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"lastModified,omitempty"`
	FetchedAt    time.Time `json:"fetchedAt"`
	LastAttempt  time.Time `json:"lastAttempt"`
	FailedAt     time.Time `json:"failedAt"`
	Size         int64     `json:"size"`
}

// Cache downloads, stores, evicts, and serves article images.
type Cache struct {
	dir    string
	client *http.Client

	mu       sync.Mutex
	entries  map[string]Entry
	total    int64
	inFlight map[string]*fetchRequest

	failureMu    sync.Mutex
	failureTimes map[string]time.Time
}

// errCacheMiss reports that a local protocol key has no prepared mapping.
var errCacheMiss = errors.New("image cache miss")

type fetchRequest struct {
	done chan struct{}
}

// Open creates a media cache below dir after clearing files left by a previous
// launch. The cache never contacts remote servers until a rewritten article
// image is requested by the WebView.
func Open(dir string) (*Cache, error) {
	if err := os.RemoveAll(dir); err != nil {
		return nil, err
	}
	cache := &Cache{
		dir:          dir,
		client:       newClient(),
		entries:      make(map[string]Entry),
		inFlight:     make(map[string]*fetchRequest),
		failureTimes: make(map[string]time.Time),
	}
	for _, name := range []string{dir, filepath.Join(dir, indexPath), filepath.Join(dir, dataPath)} {
		if err := os.MkdirAll(name, 0o700); err != nil {
			return nil, err
		}
	}
	if err := cache.load(); err != nil {
		return nil, err
	}
	return cache, nil
}

// Close releases cache resources. Downloads use short-lived files and clients,
// so no background worker has to be stopped.
func (c *Cache) Close() error {
	return nil
}

// Handler returns the handler to register with mygo.Protocol.
func (c *Cache) Handler() http.Handler {
	return http.HandlerFunc(c.ServeHTTP)
}

// load reads existing metadata and removes orphaned image files.
func (c *Cache) load() error {
	items, err := os.ReadDir(filepath.Join(c.dir, indexPath))
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, item := range items {
		if item.IsDir() || !strings.HasSuffix(item.Name(), ".json") {
			continue
		}
		key := strings.TrimSuffix(item.Name(), ".json")
		if !validKey(key) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(c.dir, indexPath, item.Name()))
		if err != nil {
			return err
		}
		var entry Entry
		if err := json.Unmarshal(data, &entry); err != nil || entry.Key != key ||
			validateSource(entry.SourceURL) != nil {
			_ = os.Remove(filepath.Join(c.dir, indexPath, item.Name()))
			_ = os.Remove(c.dataFile(key))
			continue
		}
		if entry.Size > 0 {
			info, err := os.Stat(c.dataFile(key))
			if err != nil || info.Size() != entry.Size {
				entry.Size = 0
				entry.FetchedAt = time.Time{}
				if err := saveEntry(c.dir, entry); err != nil {
					return err
				}
				_ = os.Remove(c.dataFile(key))
			} else {
				c.total += entry.Size
			}
		}
		c.entries[key] = entry
	}

	binaryItems, err := os.ReadDir(filepath.Join(c.dir, dataPath))
	if err != nil {
		return err
	}
	for _, item := range binaryItems {
		if item.IsDir() {
			continue
		}
		key := strings.TrimSuffix(item.Name(), dataSuffix)
		_, cached := c.entries[key]
		entry := c.entries[key]
		if !cached || entry.Size == 0 {
			_ = os.Remove(c.dataFile(key))
		}
	}
	return nil
}

// ServeHTTP answers local image requests from rewritten article HTML.
func (c *Cache) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	started := time.Now()
	status, cause := c.serveHTTP(writer, request, started)
	if status == http.StatusServiceUnavailable && request.Context().Err() == nil &&
		c.reportFailure(strings.TrimPrefix(request.URL.Path, "/i/"), started) {
		args := []any{
			"operation", "media.serveHTTP",
			"cacheKey", strings.TrimPrefix(request.URL.Path, "/i/"),
			"elapsedMs", time.Since(started).Milliseconds(),
		}
		if cause != nil {
			args = append(args, "error", diagnosticError(cause))
		}
		slog.Warn("media image unavailable", args...)
	}
}

// serveHTTP answers one media request and reports the response status.
func (c *Cache) serveHTTP(
	writer http.ResponseWriter,
	request *http.Request,
	started time.Time,
) (int, error) {
	if request.Method != http.MethodGet && request.Method != http.MethodHead ||
		!strings.EqualFold(request.URL.Host, "localhost") {
		http.Error(writer, "not found", http.StatusNotFound)
		return http.StatusNotFound, errors.New("invalid media protocol method or host")
	}

	prefix := "/i/"
	if !strings.HasPrefix(request.URL.Path, prefix) {
		http.Error(writer, "not found", http.StatusNotFound)
		return http.StatusNotFound, errors.New("invalid media protocol path")
	}
	key := strings.TrimPrefix(request.URL.Path, prefix)
	if !validKey(key) {
		http.Error(writer, "not found", http.StatusNotFound)
		return http.StatusNotFound, errors.New("invalid media protocol key")
	}

	entry, err := c.get(request.Context(), key)
	if err != nil {
		if entry.Size > 0 {
			if request.Context().Err() == nil && c.reportFailure(key, started) {
				slog.Warn("stale image served after refresh failure",
					"operation", "media.serveHTTP",
					"cacheKey", key,
					"error", diagnosticError(err),
				)
			}
		} else {
			http.Error(writer, "image unavailable", http.StatusServiceUnavailable)
			return http.StatusServiceUnavailable, err
		}
	}

	file, err := os.Open(c.dataFile(key))
	if err != nil {
		http.Error(writer, "image unavailable", http.StatusServiceUnavailable)
		return http.StatusServiceUnavailable, fmt.Errorf("could not open image bytes: %w", err)
	}
	defer file.Close()

	writer.Header().Set("Content-Type", entry.ContentType)
	writer.Header().Set("Cache-Control", "private, max-age=300")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeContent(writer, request, "", entry.FetchedAt, file)
	return http.StatusOK, nil
}

// reportFailure keeps repeated failures for the same image from filling the
// local diagnostic log.
func (c *Cache) reportFailure(key string, failedAt time.Time) bool {
	c.failureMu.Lock()
	defer c.failureMu.Unlock()

	if last, ok := c.failureTimes[key]; ok && failedAt.Sub(last) < time.Minute {
		return false
	}
	if len(c.failureTimes) >= maxFailures {
		return false
	}
	c.failureTimes[key] = failedAt
	return true
}

// get returns a fresh cached image, revalidates a stale one, downloads a
// prepared one, or reports why no usable bytes are available. A stale image is
// returned together with the refresh error.
func (c *Cache) get(ctx context.Context, key string) (Entry, error) {
	return c.getOnce(ctx, key, 0)
}

func (c *Cache) getOnce(ctx context.Context, key string, depth int) (Entry, error) {
	if depth > maxRedirect {
		return Entry{}, errors.New("too many image cache redirects")
	}

	c.mu.Lock()
	entry, ok := c.entries[key]
	if !ok {
		c.mu.Unlock()
		return Entry{}, errCacheMiss
	}
	if c.imageFresh(entry) {
		entry.LastAttempt = time.Now()
		c.entries[key] = entry
		c.mu.Unlock()
		return entry, nil
	}
	if imageNegative(entry) && entry.Size == 0 {
		c.mu.Unlock()
		return Entry{}, errors.New("the image recently failed to download")
	}
	if existing, waiting := c.inFlight[key]; waiting {
		c.mu.Unlock()
		<-existing.done
		return c.getOnce(ctx, key, depth+1)
	}
	current := &fetchRequest{done: make(chan struct{})}
	c.inFlight[key] = current
	c.mu.Unlock()

	fetched, err := c.fetch(ctx, entry)
	if err != nil && entry.Size > 0 {
		c.mu.Lock()
		delete(c.inFlight, key)
		close(current.done)
		c.mu.Unlock()
		entry.FailedAt = time.Now()
		entry.LastAttempt = entry.FailedAt
		return entry, err
	}
	c.mu.Lock()
	delete(c.inFlight, key)
	close(current.done)
	c.mu.Unlock()
	if err != nil {
		return Entry{}, err
	}
	return fetched, nil
}

// imageFresh reports whether usable image bytes are still current enough.
func (c *Cache) imageFresh(entry Entry) bool {
	return entry.Size > 0 && entry.FetchedAt.Add(freshTTL).After(time.Now())
}

// imageNegative reports whether a recent failed download should not be tried
// again.
func imageNegative(entry Entry) bool {
	return entry.Size == 0 && !entry.FailedAt.IsZero() &&
		entry.FailedAt.Add(failureTTL).After(time.Now())
}

// fetch downloads or revalidates one image and stores the accepted bytes.
func (c *Cache) fetch(ctx context.Context, old Entry) (Entry, error) {
	now := time.Now()
	entry := old
	entry.LastAttempt = now

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, old.SourceURL, nil)
	if err != nil {
		c.fail(old, now)
		return Entry{}, err
	}
	request.Header.Set("User-Agent", userAgent)
	request.Header.Set("Accept", imageAccept)
	if old.ETag != "" {
		request.Header.Set("If-None-Match", old.ETag)
	}
	if old.LastModified != "" {
		request.Header.Set("If-Modified-Since", old.LastModified)
	}

	response, err := c.client.Do(request)
	if err != nil {
		if request.Context().Err() != nil {
			return Entry{}, err
		}
		c.fail(old, now)
		return Entry{}, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotModified && old.Size > 0 {
		entry.ContentType = old.ContentType
		entry.FetchedAt = now
		entry.FailedAt = time.Time{}
		return entry, c.put(old.Key, entry, nil)
	}
	if response.StatusCode != http.StatusOK {
		c.fail(old, now)
		return Entry{}, fmt.Errorf("the image server answered %s", response.Status)
	}
	if response.ContentLength > MaxBytes {
		c.fail(old, now)
		return Entry{}, errors.New("the image is larger than the local limit")
	}

	data, err := io.ReadAll(io.LimitReader(response.Body, MaxBytes+1))
	if err != nil {
		if request.Context().Err() != nil {
			return Entry{}, err
		}
		c.fail(old, now)
		return Entry{}, err
	}
	if len(data) > MaxBytes {
		c.fail(old, now)
		return Entry{}, errors.New("the image is larger than the local limit")
	}

	contentType := imageType(data, response.Header.Get("Content-Type"))
	if contentType == "" {
		c.fail(old, now)
		return Entry{}, errors.New("the address does not serve a supported image")
	}

	entry.ContentType = contentType
	entry.ETag = response.Header.Get("ETag")
	entry.LastModified = response.Header.Get("Last-Modified")
	entry.FetchedAt = now
	entry.FailedAt = time.Time{}
	entry.Size = int64(len(data))
	return entry, c.put(old.Key, entry, data)
}

// diagnosticError removes image source URLs from an error before logging it.
func diagnosticError(err error) string {
	return loggedURLPattern.ReplaceAllString(err.Error(), "${1}[redacted]")
}

// fail records a short-lived negative result while retaining stale bytes.
func (c *Cache) fail(old Entry, failedAt time.Time) {
	entry := old
	entry.LastAttempt = failedAt
	entry.FailedAt = failedAt
	_ = c.put(old.Key, entry, nil)
}

// put atomically installs new image bytes and metadata, then enforces the
// total cache limit.
func (c *Cache) put(key string, entry Entry, data []byte) error {
	if data != nil {
		if err := writeData(c.dataFile(key), data); err != nil {
			return err
		}
	}
	if err := saveEntry(c.dir, entry); err != nil {
		_ = os.Remove(c.dataFile(key))
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	old, existed := c.entries[key]
	if existed && old.Size > 0 {
		c.total -= old.Size
	}
	c.entries[key] = entry
	c.total += entry.Size
	c.evictLocked(key)
	return nil
}

// evictLocked removes least recently used images until the cache fits its
// limit. The key just stored is never evicted.
func (c *Cache) evictLocked(keep string) {
	if c.total <= MaxCacheBytes {
		return
	}
	keys := make([]string, 0, len(c.entries))
	for key, entry := range c.entries {
		if key != keep && entry.Size > 0 {
			keys = append(keys, key)
		}
	}
	slices.SortFunc(keys, func(first, second string) int {
		return c.entries[first].LastAttempt.Compare(c.entries[second].LastAttempt)
	})
	for _, key := range keys {
		if c.total <= MaxCacheBytes {
			return
		}
		entry := c.entries[key]
		_ = os.Remove(c.dataFile(key))
		_ = os.Remove(filepath.Join(c.dir, indexPath, key+".json"))
		delete(c.entries, key)
		c.total -= entry.Size
	}
}

// prepare records the mapping needed to resolve a rewritten URL later.
func (c *Cache) prepare(key, sourceURL string) bool {
	c.mu.Lock()
	if _, ok := c.entries[key]; ok {
		c.mu.Unlock()
		return true
	}
	if len(c.entries) >= MaxEntries {
		c.mu.Unlock()
		return false
	}

	entry := Entry{Key: key, SourceURL: sourceURL}
	c.mu.Unlock()
	if err := saveEntry(c.dir, entry); err != nil {
		slog.Warn("image mapping save failed",
			"operation", "media.prepare", "cacheKey", key, "error", err)
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if current, ok := c.entries[key]; ok {
		entry = current
	} else {
		c.entries[key] = entry
	}
	return true
}

func validKey(key string) bool {
	if len(key) != sha256.Size*2 {
		return false
	}
	for _, char := range key {
		if !('0' <= char && char <= '9' || 'a' <= char && char <= 'f') {
			return false
		}
	}
	return true
}

func keyFor(sourceURL string) string {
	digest := sha256.Sum256([]byte(sourceURL))
	return hex.EncodeToString(digest[:])
}

func (c *Cache) dataFile(key string) string {
	return filepath.Join(c.dir, dataPath, key+dataSuffix)
}

func entryFile(dir, key string) string {
	return filepath.Join(dir, indexPath, key+".json")
}

func saveEntry(dir string, entry Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return writeAtomic(entryFile(dir, entry.Key), data, 0o600)
}

func writeData(path string, data []byte) error {
	return writeAtomic(path, data, 0o600)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)

	if _, err = file.Write(data); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	if err = os.Chmod(temp, mode); err != nil {
		return err
	}
	return os.Rename(temp, path)
}
