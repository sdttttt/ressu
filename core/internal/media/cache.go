// Package media caches article images and serves them to the WebView through
// a local MyGO protocol.
package media

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Scheme is the local protocol used by rewritten article images.
const Scheme = "ressu-media"

// MaxBytes bounds a single downloaded image.
const MaxBytes = 8 << 20

const (
	// MinCacheBytes bounds the disk cache from below.
	MinCacheBytes = 32 << 20

	// DefaultCacheBytes is the disk cache limit of a new install.
	DefaultCacheBytes = 256 << 20

	// MaxCacheBytes bounds the disk cache from above.
	MaxCacheBytes = 512 << 20
)

// MaxEntries bounds source mappings created for one local cache.
const MaxEntries = 10000

const (
	freshTTL   = 7 * 24 * time.Hour
	failureTTL = time.Hour
)

// Entry describes one cached image. A nil size marks a prepared, but not yet
// downloaded resource.
type Entry struct {
	Key         string `json:"key"`
	SourceURL   string `json:"sourceUrl"`
	ContentType string `json:"contentType"`
	// CreatedAt is when the cached image bytes were first stored.
	CreatedAt    time.Time `json:"createdAt"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"lastModified,omitempty"`
	FetchedAt    time.Time `json:"fetchedAt"`
	LastAttempt  time.Time `json:"lastAttempt"`
	FailedAt     time.Time `json:"failedAt"`
	Size         int64     `json:"size"`
}

// Cache downloads, stores, evicts, and serves article images.
type Cache struct {
	dir      string
	client   *http.Client
	maxBytes int64
	openedAt time.Time

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

type Options struct {
	// MaxBytes bounds image bytes stored on disk. Zero uses the default.
	MaxBytes int64
}

// Open opens a persistent image cache, loads its mappings, and removes the
// oldest bytes that exceed the limit. The cache never contacts remote servers
// until a rewritten article image is requested by the WebView.
func Open(dir string, options Options) (*Cache, error) {
	for _, name := range []string{dir, filepath.Join(dir, indexPath), filepath.Join(dir, dataPath)} {
		if err := os.MkdirAll(name, 0o700); err != nil {
			return nil, err
		}
	}
	cache := &Cache{
		dir:          dir,
		client:       newClient(),
		maxBytes:     normalizeLimit(options.MaxBytes),
		openedAt:     time.Now(),
		entries:      make(map[string]Entry),
		inFlight:     make(map[string]*fetchRequest),
		failureTimes: make(map[string]time.Time),
	}
	if err := cache.load(); err != nil {
		return nil, err
	}
	if err := cache.enforceLimitLocked(); err != nil {
		return nil, err
	}
	return cache, nil
}

// SetLimit changes the disk cache limit and removes the oldest bytes that now
// exceed the limit.
func (c *Cache) SetLimit(maxBytes int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxBytes = normalizeLimit(maxBytes)
	return c.enforceLimitLocked()
}

// Usage returns the bytes currently stored in the cache and the active byte
// limit.
func (c *Cache) Usage() (used int64, limit int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.total, c.maxBytes
}

// Get resolves a data URL, a local protocol URL, or a public remote image
// URL to its cache entry. For remote images it first maps the source to this
// cache's local URL. A usable stale entry is returned even when refreshing
// it failed; read its bytes with Read.
func (c *Cache) Get(ctx context.Context, source string) (Entry, error) {
	if c == nil {
		return Entry{}, ErrImageUnavailable
	}
	source = strings.TrimSpace(source)
	if strings.HasPrefix(strings.ToLower(source), "data:") {
		data, err := decodeDataURL(source)
		if err != nil {
			return Entry{}, err
		}
		return Entry{
			Key:         keyFor(source),
			SourceURL:   source,
			ContentType: imageType(data, ""),
			CreatedAt:   time.Now(),
			FetchedAt:   time.Now(),
			LastAttempt: time.Now(),
			Size:        int64(len(data)),
		}, nil
	}

	key, ok := localImageKey(source)
	if ok {
		return c.getUsable(ctx, key)
	}

	viewerURL, ok := c.LocalURL(source)
	if !ok {
		return Entry{}, ErrInvalidImageSource
	}
	parsed, err := url.Parse(viewerURL)
	if err != nil {
		return Entry{}, ErrInvalidImageSource
	}
	key, ok = localImageKey(viewerURL)
	if !ok || parsed.RawQuery != "" || parsed.Fragment != "" {
		return Entry{}, ErrInvalidImageSource
	}
	return c.getUsable(ctx, key)
}

// getUsable returns a stale entry when a refresh failed but older bytes
// remain available.
func (c *Cache) getUsable(ctx context.Context, key string) (Entry, error) {
	entry, getErr := c.get(ctx, key)
	if getErr != nil && entry.Size == 0 {
		return Entry{}, getErr
	}
	return entry, nil
}

// Close releases cache resources. Downloads use short-lived files and clients,
// so no background worker has to be stopped.
func (c *Cache) Close() error {
	return nil
}

// Read returns the image bytes described by an entry returned from Get. A
// stale image is still readable when its refresh failed.
func (c *Cache) Read(entry Entry) ([]byte, error) {
	if strings.HasPrefix(strings.ToLower(entry.SourceURL), "data:") {
		return decodeDataURL(entry.SourceURL)
	}

	if !validKey(entry.Key) {
		return nil, ErrImageUnavailable
	}

	data, err := os.ReadFile(c.dataFile(entry.Key))
	if err != nil || len(data) == 0 || entry.Size > 0 && int64(len(data)) != entry.Size {
		return nil, ErrImageUnavailable
	}
	return data, nil
}

func (c *Cache) get(ctx context.Context, key string) (Entry, error) {
	return cacheGet(c, ctx, key)
}

func (c *Cache) getOnce(ctx context.Context, key string, depth int) (Entry, error) {
	return cacheGetOnce(c, ctx, key, depth)
}

func (c *Cache) imageFresh(entry Entry) bool {
	return cacheImageFresh(entry)
}

// ViewerURL returns the local protocol URL for a cache entry. Data URLs are
// already safe to load directly.
func (e Entry) ViewerURL() string {
	if strings.HasPrefix(strings.ToLower(e.SourceURL), "data:") {
		return e.SourceURL
	}
	return localURL(e.Key)
}

func (c *Cache) fetch(ctx context.Context, old Entry) (Entry, error) {
	return cacheFetch(c, ctx, old)
}

func (c *Cache) fail(old Entry, failedAt time.Time) {
	cacheFail(c, old, failedAt)
}

// Handler returns the handler to register with mygo.Protocol.
func (c *Cache) Handler() http.Handler {
	return http.HandlerFunc(c.ServeHTTP)
}

// ServeHTTP answers local image requests from rewritten article HTML.
func (c *Cache) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	cacheServeHTTP(c, writer, request)
}

func (c *Cache) serveHTTP(
	writer http.ResponseWriter,
	request *http.Request,
	started time.Time,
) (int, error) {
	return cacheServeRequest(c, writer, request, started)
}

func (c *Cache) reportFailure(key string, failedAt time.Time) bool {
	return cacheReportFailure(c, key, failedAt)
}

func (c *Cache) load() error {
	return cacheLoad(c)
}

func (c *Cache) put(key string, entry Entry, data []byte) error {
	return cachePut(c, key, entry, data)
}

func (c *Cache) evictLocked(keep string) error {
	return cacheEvictLocked(c, keep)
}

func (c *Cache) enforceLimitLocked() error {
	return cacheEnforceLimitLocked(c)
}

func (c *Cache) prepare(key, sourceURL string) bool {
	return cachePrepare(c, key, sourceURL)
}

func (c *Cache) dataFile(key string) string {
	return cacheDataFile(c, key)
}

// LocalURL maps one public image address to its stable local protocol URL. It
// persists only the source mapping; bytes are downloaded when WebView asks.
func (c *Cache) LocalURL(raw string) (string, bool) {
	return cacheLocalURL(c, raw)
}

// RewriteHTML replaces remote img addresses with local cache URLs. Images that
// cannot be mapped lose their src instead of remaining reachable directly.
func (c *Cache) RewriteHTML(markup, baseURL string) string {
	return cacheRewriteHTML(c, markup, baseURL)
}
