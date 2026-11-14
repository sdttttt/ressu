package media

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSetLimitPrunesOldestImages(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "media")
	cache, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	oldData := bytes.Repeat([]byte("old"), MinCacheBytes/3+1)
	oldKey := newTestCacheEntry(t, dir, "https://cdn.example/old.png", oldData, time.Now().Add(-time.Minute))
	newKey := newTestCacheEntry(t, dir, "https://cdn.example/new.png", []byte("new!"), time.Now())
	cache.mu.Lock()
	cache.entries[oldKey] = Entry{
		Key:       oldKey,
		CreatedAt: time.Now().Add(-time.Minute),
		Size:      int64(len(oldData)),
	}
	cache.entries[newKey] = Entry{Key: newKey, CreatedAt: time.Now(), Size: 4}
	cache.total = int64(len(oldData)) + 4
	cache.mu.Unlock()

	if err := cache.SetLimit(MinCacheBytes); err != nil {
		t.Fatal(err)
	}

	cache.mu.Lock()
	_, oldCached := cache.entries[oldKey]
	_, newCached := cache.entries[newKey]
	total := cache.total
	maxBytes := cache.maxBytes
	cache.mu.Unlock()

	if oldCached || !newCached || total != 4 || maxBytes != MinCacheBytes {
		t.Fatalf("old=%t new=%t total=%d max=%d; want old removed and new retained", oldCached, newCached, total, maxBytes)
	}
}

func TestUsageReflectsStoredBytesAndLimit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "media")
	cache, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	used, limit := cache.Usage()
	if used != 0 || limit != DefaultCacheBytes {
		t.Fatalf("Usage() = %d/%d, want 0/%d", used, limit, DefaultCacheBytes)
	}

	cache.mu.Lock()
	cache.entries["a"] = Entry{Key: "a", CreatedAt: time.Now(), Size: 100}
	cache.entries["b"] = Entry{Key: "b", CreatedAt: time.Now(), Size: 23}
	cache.total = 123
	cache.mu.Unlock()

	used, limit = cache.Usage()
	if used != 123 || limit != DefaultCacheBytes {
		t.Fatalf("Usage() = %d/%d, want 123/%d", used, limit, DefaultCacheBytes)
	}
}

func newTestCacheEntry(t *testing.T, dir, source string, data []byte, createdAt time.Time) string {
	t.Helper()
	for _, name := range []string{indexPath, dataPath} {
		if err := os.MkdirAll(filepath.Join(dir, name), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	key := keyFor(source)
	entry := Entry{
		Key:       key,
		SourceURL: source,
		CreatedAt: createdAt,
		FetchedAt: createdAt,
		Size:      int64(len(data)),
	}
	if err := os.WriteFile(filepath.Join(dir, dataPath, key+dataSuffix), data, 0o600); err != nil {
		t.Fatal(err)
	}
	metadata, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(entryPath(dir, key), metadata, 0o600); err != nil {
		t.Fatal(err)
	}
	return key
}
