package media

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/sdttttt/ressu/core/internal/logging"
)

const (
	indexPath  = "index"
	dataPath   = "data"
	dataSuffix = ".bin"
)

// cacheLoad reads existing metadata and removes orphaned image files.
func cacheLoad(c *Cache) error {
	items, err := os.ReadDir(filepath.Join(c.dir, indexPath))
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, item := range items {
		if item.IsDir() || strings.HasPrefix(item.Name(), ".") {
			if strings.HasPrefix(item.Name(), ".") {
				_ = os.Remove(filepath.Join(c.dir, indexPath, item.Name()))
			}
			continue
		}
		key := strings.TrimSuffix(item.Name(), ".json")
		if !validKey(key) {
			_ = os.Remove(filepath.Join(c.dir, indexPath, item.Name()))
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
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = entry.FetchedAt
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = entry.LastAttempt
		}
		if entry.CreatedAt.IsZero() {
			entry.CreatedAt = time.Now()
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
		if strings.HasPrefix(item.Name(), ".") {
			_ = os.Remove(filepath.Join(c.dir, dataPath, item.Name()))
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

// cachePut atomically installs new image bytes and metadata, then enforces the
// total cache limit.
func cachePut(c *Cache, key string, entry Entry, data []byte) error {
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
	return c.evictLocked(key)
}

// cacheEvictLocked removes oldest images until the cache fits its limit. The
// key just stored is never evicted.
func cacheEvictLocked(c *Cache, keep string) error {
	if c.total <= c.maxBytes {
		return nil
	}
	keys := make([]string, 0, len(c.entries))
	for key, entry := range c.entries {
		if key != keep && entry.Size > 0 {
			keys = append(keys, key)
		}
	}
	slices.SortFunc(keys, func(first, second string) int {
		return c.entries[first].CreatedAt.Compare(c.entries[second].CreatedAt)
	})
	var (
		evicted int
		freed   int64
	)
	for _, key := range keys {
		if c.total <= c.maxBytes {
			break
		}
		entry := c.entries[key]
		if err := removeEntry(c.dir, key); err != nil {
			return err
		}
		delete(c.entries, key)
		c.total -= entry.Size
		evicted++
		freed += entry.Size
	}
	if evicted > 0 {
		logging.Debug("media cache evicted",
			"operation", "media.evict",
			"entries", evicted,
			"freedBytes", freed,
		)
	}
	return nil
}

// cacheEnforceLimitLocked checks the whole loaded cache at startup and after
// the preference changes.
func cacheEnforceLimitLocked(c *Cache) error {
	return cacheEvictLocked(c, "")
}

// cachePrepare records the mapping needed to resolve a rewritten URL later.
func cachePrepare(c *Cache, key, sourceURL string) bool {
	c.mu.Lock()
	if _, ok := c.entries[key]; ok {
		c.mu.Unlock()
		return true
	}
	if len(c.entries) >= MaxEntries {
		c.mu.Unlock()
		return false
	}

	entry := Entry{Key: key, SourceURL: sourceURL, CreatedAt: time.Now()}
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

func cacheDataFile(c *Cache, key string) string {
	return filepath.Join(c.dir, dataPath, key+dataSuffix)
}

func entryPath(dir, key string) string {
	return filepath.Join(dir, indexPath, key+".json")
}

func removeEntry(dir, key string) error {
	paths := []string{filepath.Join(dir, dataPath, key+dataSuffix), entryPath(dir, key)}
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("could not remove image cache entry: %w", err)
		}
	}
	return nil
}

func saveEntry(dir string, entry Entry) error {
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return writeAtomic(entryPath(dir, entry.Key), data, 0o600)
}

func normalizeLimit(maxBytes int64) int64 {
	switch {
	case maxBytes <= 0:
		return DefaultCacheBytes
	case maxBytes < MinCacheBytes:
		return MinCacheBytes
	case maxBytes > MaxCacheBytes:
		return MaxCacheBytes
	default:
		return maxBytes
	}
}

func writeData(path string, data []byte) error {
	return writeAtomic(path, data, 0o600)
}

func writeAtomic(path string, data []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".")
	if err != nil {
		return err
	}
	temp := file.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(temp)
		}
	}()

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
	if err = os.Rename(temp, path); err != nil {
		return err
	}
	removeTemp = false
	return nil
}
