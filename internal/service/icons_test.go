package service

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/store"
)

var iconPNG = func() []byte {
	data, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
	)
	if err != nil {
		panic(err)
	}
	return data
}()

func iconFeedXML(iconURL string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel>
<title>图标源</title>
<link>https://site.example/</link>
<image><url>` + iconURL + `</url><title>图标</title><link>https://site.example/</link></image>
</channel></rss>`
}

func TestFeedsIconUsesAndReusesLocalCache(t *testing.T) {
	var hits atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/icon.png" {
			hits.Add(1)
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(iconPNG)
			return
		}
		_, _ = io.WriteString(w, iconFeedXML(server.URL+"/icon.png"))
	}))
	t.Cleanup(server.Close)

	feeds := NewFeeds(openStore(t))
	feed, err := feeds.Add(context.Background(), "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if feed.IconURL != server.URL+"/icon.png" {
		t.Fatalf("icon URL = %q", feed.IconURL)
	}

	icon, err := feeds.Icon(context.Background(), feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if icon.Meta.SourceURL != feed.IconURL || icon.Meta.ContentType != "image/png" ||
		string(icon.Data) != string(iconPNG) {
		t.Fatalf("icon = %+v", icon)
	}

	if _, err = feeds.Icon(context.Background(), feed.ID); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Errorf("icon requests = %d, want 1", hits.Load())
	}

	if err = feeds.store.DeleteIcon(feed.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = feeds.Icon(context.Background(), feed.ID); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 2 {
		t.Errorf("icon requests after delete = %d, want 2", hits.Load())
	}
}

func TestFeedsIconInvalidatesAChangedSource(t *testing.T) {
	var hits atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/new-icon.png" {
			hits.Add(1)
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(iconPNG)
			return
		}
		_, _ = io.WriteString(w, iconFeedXML(server.URL+"/new-icon.png"))
	}))
	t.Cleanup(server.Close)

	feeds := NewFeeds(openStore(t))
	feed, err := feeds.Add(context.Background(), "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	old := store.Icon{
		Meta: store.IconMeta{
			SourceURL:   server.URL + "/old-icon.png",
			ContentType: "image/png",
			FetchedAt:   time.Now(),
		},
		Data: []byte("old"),
	}
	if err = feeds.store.PutIcon(feed.ID, old); err != nil {
		t.Fatal(err)
	}

	icon, err := feeds.Icon(context.Background(), feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if icon.Meta.SourceURL != server.URL+"/new-icon.png" || hits.Load() != 1 {
		t.Fatalf("icon = %+v, hits = %d", icon, hits.Load())
	}
}

func TestFeedsIconExpiresAfterTTL(t *testing.T) {
	var hits atomic.Int32
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/icon.png" {
			hits.Add(1)
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(iconPNG)
			return
		}
		_, _ = io.WriteString(w, iconFeedXML(server.URL+"/icon.png"))
	}))
	t.Cleanup(server.Close)

	feeds := NewFeeds(openStore(t))
	feed, err := feeds.Add(context.Background(), "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	cached := store.Icon{
		Meta: store.IconMeta{
			SourceURL:   feed.IconURL,
			ContentType: "image/png",
			FetchedAt:   time.Now().Add(-iconTTL - time.Minute),
		},
		Data: []byte("old"),
	}
	if err = feeds.store.PutIcon(feed.ID, cached); err != nil {
		t.Fatal(err)
	}

	if _, err = feeds.Icon(context.Background(), feed.ID); err != nil {
		t.Fatal(err)
	}
	if hits.Load() != 1 {
		t.Errorf("icon requests = %d, want 1", hits.Load())
	}
}

func TestFeedsIconSharesOneRequest(t *testing.T) {
	var hits atomic.Int32
	release := make(chan struct{})
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/blocked.png" {
			if hits.Add(1) == 1 {
				<-release
			}
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write(iconPNG)
			return
		}
		_, _ = io.WriteString(w, iconFeedXML(""))
	}))
	t.Cleanup(server.Close)

	go func() {
		time.Sleep(20 * time.Millisecond)
		close(release)
	}()

	feeds := NewFeeds(openStore(t))
	feed, err := feeds.Add(context.Background(), "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	feed.IconURL = server.URL + "/blocked.png"
	feed.SiteURL = "not a url"
	if err = feeds.store.UpdateFeed(feed); err != nil {
		t.Fatal(err)
	}

	const calls = 2
	icons := make([]store.Icon, calls)
	errs := make([]error, calls)
	var wg sync.WaitGroup
	for i := range calls {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			icons[i], errs[i] = feeds.Icon(context.Background(), feed.ID)
		}(i)
	}
	wg.Wait()

	for i := range calls {
		if errs[i] != nil {
			t.Fatalf("call %d error = %v", i, errs[i])
		}
		if string(icons[i].Data) != string(iconPNG) {
			t.Fatalf("call %d icon = %+v", i, icons[i])
		}
	}
	if hits.Load() != 1 {
		t.Errorf("icon requests = %d, want 1", hits.Load())
	}
}
