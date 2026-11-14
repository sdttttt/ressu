package service

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sdttttt/ressu/core/internal/store"
)

// newTestFeed gives the refresh test a stored address without network setup.
func newTestFeed(serverURL string) store.Feed {
	return store.Feed{URL: serverURL, Title: "Test Feed"}
}

func TestRefreshFailureIsLoggedWithFeedContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, "boom", http.StatusInternalServerError)
	}))
	defer server.Close()

	st := openStore(t)
	feed, err := st.AddFeed(newTestFeed(server.URL))
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	feeds := NewFeeds(st)
	if _, err := feeds.Refresh(context.Background(), feed.ID); err == nil {
		t.Fatal("the refresh should fail")
	}

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("log entry is not JSON: %v", err)
	}
	if entry["level"] != "WARN" || entry["operation"] != "feeds.refresh" ||
		entry["feedId"] != float64(feed.ID) || entry["elapsedMs"] == nil {
		t.Fatalf("unexpected refresh log: %#v", entry)
	}
	if text := output.String(); strings.Contains(text, server.URL+"/other") {
		t.Fatalf("log leaked an unrelated address: %s", text)
	}
}
