//go:build debug

package logging

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"
)

func TestDebugLogsWithTag(t *testing.T) {
	file, err := Open(t.TempDir(), Options{MaxBytes: -1})
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	slog.SetDefault(file.Logger())

	Debug("debug diagnostic", "operation", "media.test")
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(file.Path())
	if err != nil {
		t.Fatal(err)
	}
	var entry map[string]any
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("debug entry is not JSON: %v", err)
	}
	if entry["level"] != "DEBUG" || entry["msg"] != "debug diagnostic" ||
		entry["operation"] != "media.test" {
		t.Fatalf("unexpected debug entry: %#v", entry)
	}
}
