package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenWritesStructuredErrors(t *testing.T) {
	file, err := Open(t.TempDir(), Options{MaxBytes: -1})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	file.Logger().Error("operation failed", "operation", "refresh", "feedId", int64(7))
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file.Path())
	if err != nil {
		t.Fatal(err)
	}
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("log entry is not JSON: %v", err)
	}
	if entry["level"] != "ERROR" || entry["msg"] != "operation failed" ||
		entry["operation"] != "refresh" || entry["feedId"] != float64(7) {
		t.Fatalf("unexpected log entry: %#v", entry)
	}
	if info, err := os.Stat(file.Path()); err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("file mode = %v, err = %v; want 0600", info.Mode(), err)
	}
}

func TestOpenFiltersBelowLevel(t *testing.T) {
	file, err := Open(t.TempDir(), Options{
		MaxBytes: -1,
		Level:    slog.LevelWarn,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	file.Logger().Info("not diagnostic")
	file.Logger().Warn("diagnostic")
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file.Path())
	if err != nil {
		t.Fatal(err)
	}
	if text := string(data); strings.Contains(text, "not diagnostic") ||
		!strings.Contains(text, "diagnostic") {
		t.Fatalf("unexpected filtered log: %s", text)
	}
}

func TestOpenRotatesWhenFull(t *testing.T) {
	dir := t.TempDir()
	file, err := Open(dir, Options{MaxBytes: 32})
	if err != nil {
		t.Fatal(err)
	}
	file.Logger().Warn(strings.Repeat("first", 10))
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	file, err = Open(dir, Options{MaxBytes: 32})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	file.Logger().Warn(strings.Repeat("second", 10))

	backup, err := os.ReadFile(filepath.Join(dir, rotatedFileName))
	if err != nil || !bytes.Contains(backup, []byte("first")) {
		t.Fatalf("rotated log = %q, err = %v", backup, err)
	}
	active, err := os.ReadFile(file.Path())
	if err != nil || !bytes.Contains(active, []byte("second")) {
		t.Fatalf("active log = %q, err = %v", active, err)
	}
}

func TestOpenPrunesOldBackups(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, FileName)
	contents := map[string]string{
		active:           "active",
		active + ".1":    "first",
		active + ".2":    "second",
		active + ".3":    "third",
		active + ".4":    "fourth",
		active + ".5":    "fifth",
		active + ".6":    "sixth",
		active + ".text": "ignored",
	}
	for path, content := range contents {
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	file, err := Open(dir, Options{MaxBytes: -1})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	expected := map[int]string{
		1: "first",
		2: "second",
		3: "third",
		4: "fourth",
	}
	for generation, want := range expected {
		path := fmt.Sprintf("%s.%d", active, generation)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != want {
			t.Fatalf("%s = %q, want %q", path, data, want)
		}
	}
	for _, name := range []string{active + ".5", active + ".6"} {
		if _, err := os.Stat(name); !os.IsNotExist(err) {
			t.Fatalf("%s still exists", name)
		}
	}
}
