package media

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenClearsPreviousLaunchCache(t *testing.T) {
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

	cache, err := Open(dir)
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
	if _, err := os.Stat(filepath.Join(dir, "root.bin")); !os.IsNotExist(err) {
		t.Fatalf("root leftover stat err = %v, want not exist", err)
	}
}
