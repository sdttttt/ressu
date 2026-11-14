package store

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	bolt "go.etcd.io/bbolt"
)

func captureSlog(t *testing.T) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &output
}

func logEntries(t *testing.T, output *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(output.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("log entry is not JSON: %v", err)
		}
		entries = append(entries, entry)
	}
	return entries
}

func findLogEntry(t *testing.T, entries []map[string]any, message string) map[string]any {
	t.Helper()
	for _, entry := range entries {
		if entry["msg"] == message {
			return entry
		}
	}
	t.Fatalf("no log entry %q in %#v", message, entries)
	return nil
}

func TestOpenLogsFeedCount(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddFeed(Feed{URL: "https://example.com/feed.xml", Title: "Test"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	output := captureSlog(t)
	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	entry := findLogEntry(t, logEntries(t, output), "database opened")
	if entry["operation"] != "store.open" || entry["feeds"] != float64(1) {
		t.Fatalf("unexpected open log: %#v", entry)
	}
}

func TestMigrateLogsUpgrade(t *testing.T) {
	dir := t.TempDir()
	db, err := bolt.Open(filepath.Join(dir, dbFileName), 0o600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		settings, err := tx.CreateBucket([]byte(settingsBucket))
		if err != nil {
			return err
		}
		return put(settings, []byte(schemaVersionKey), 2)
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	output := captureSlog(t)
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	entry := findLogEntry(t, logEntries(t, output), "database schema migrating")
	if entry["operation"] != "store.migrate" ||
		entry["from"] != float64(2) || entry["to"] != float64(currentSchemaVersion) {
		t.Fatalf("unexpected migrate log: %#v", entry)
	}
}
