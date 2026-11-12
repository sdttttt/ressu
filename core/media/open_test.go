package media

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestOpenLoadsPreviousLaunchCache(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "articles")
	if err := os.MkdirAll(filepath.Join(dir, indexPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, dataPath), 0o700); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(dir, indexPath, "stale.json"): `{}`,
		filepath.Join(dir, dataPath, "stale.bin"):   "image bytes",
		filepath.Join(dir, "root.bin"):              "leftover",
	}
	for path, data := range files {
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	cache, err := Open(dir, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	if len(cache.entries) != 0 || cache.total != 0 {
		t.Fatalf("cache = %d entries, %d bytes; want empty", len(cache.entries), cache.total)
	}
	for _, name := range []string{indexPath, dataPath} {
		items, err := os.ReadDir(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 0 {
			t.Fatalf("%s contains %d items after startup", name, len(items))
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "root.bin")); err != nil {
		t.Fatalf("unrelated root file should be left alone: %v", err)
	}
}

func TestOpenPrunesOldestImagesBeyondLimit(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "articles")
	oldKey := newTestCacheEntry(
		t,
		dir,
		"https://cdn.example/old.png",
		bytes.Repeat([]byte{1}, MinCacheBytes),
		time.Now().Add(-time.Minute),
	)
	newKey := newTestCacheEntry(t, dir, "https://cdn.example/new.png", []byte{1}, time.Now())
	if oldKey == newKey {
		t.Fatal("test entries have the same key")
	}

	cache, err := Open(dir, Options{MaxBytes: 4})
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	if _, cached := cache.entries[oldKey]; cached || cache.total != 1 {
		t.Fatalf("cache = %d bytes, old cached = %t; want old removed", cache.total, cached)
	}
	if _, cached := cache.entries[newKey]; !cached {
		t.Fatal("new image was removed")
	}
	if _, err := os.Stat(cache.dataFile(oldKey)); !os.IsNotExist(err) {
		t.Fatalf("old data stat err = %v, want not exist", err)
	}
	if _, err := os.Stat(entryPath(dir, oldKey)); !os.IsNotExist(err) {
		t.Fatalf("old metadata stat err = %v, want not exist", err)
	}
}
