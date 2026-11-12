package icon_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/sdttttt/ressu/core/icon"
)

var pngBytes = func() []byte {
	data, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==",
	)
	if err != nil {
		panic(err)
	}
	return data
}()

func iconServer(t *testing.T, hits *atomic.Int32) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/icon") {
			if hits != nil {
				hits.Add(1)
			}
			if r.Header.Get("User-Agent") == "" {
				t.Errorf("icon request has no user agent")
			}
		}

		switch r.URL.Path {
		case "/feed-icon.png", "/site-icon.png", "/favicon.ico", "/large.png":
			w.Header().Set("Content-Type", "image/png")
			if r.URL.Path == "/large.png" {
				data := bytes.Repeat(pngBytes, (icon.MaxBytes/len(pngBytes))+1)
				_, _ = w.Write(data)
				return
			}
			_, _ = w.Write(pngBytes)
		case "/site":
			w.Header().Set("Content-Type", "text/html")
			_, _ = io.WriteString(w, `<html><head><link rel="apple-touch-icon" href="/site-icon.png"></head></html>`)
		case "/bad.png":
			w.Header().Set("Content-Type", "image/svg+xml")
			_, _ = io.WriteString(w, `<svg xmlns="http://www.w3.org/2000/svg"></svg>`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestFetchPrefersTheExplicitFeedIcon(t *testing.T) {
	server := iconServer(t, nil)
	image, err := icon.Fetch(context.Background(), icon.Source{
		FeedURL: server.URL + "/feed.xml",
		SiteURL: server.URL + "/site",
		IconURL: "/feed-icon.png",
	})
	if err != nil {
		t.Fatal(err)
	}
	if image.SourceURL != server.URL+"/feed-icon.png" {
		t.Errorf("source = %q", image.SourceURL)
	}
	if image.ContentType != "image/png" || !bytes.Equal(image.Data, pngBytes) {
		t.Errorf("image = %+v", image)
	}
}

func TestFetchDiscoversASiteFaviconLink(t *testing.T) {
	server := iconServer(t, nil)
	image, err := icon.Fetch(context.Background(), icon.Source{
		FeedURL: server.URL + "/feed.xml",
		SiteURL: server.URL + "/site",
	})
	if err != nil {
		t.Fatal(err)
	}
	if image.SourceURL != server.URL+"/site-icon.png" {
		t.Errorf("source = %q", image.SourceURL)
	}
}

func TestFetchFallsBackToFaviconICO(t *testing.T) {
	server := iconServer(t, nil)
	image, err := icon.Fetch(context.Background(), icon.Source{
		FeedURL: server.URL + "/feed.xml",
		SiteURL: server.URL + "/missing-site",
	})
	if err != nil {
		t.Fatal(err)
	}
	if image.SourceURL != server.URL+"/favicon.ico" {
		t.Errorf("source = %q", image.SourceURL)
	}
}

func TestFetchRejectsOversizedAndUnsupportedImages(t *testing.T) {
	var hits atomic.Int32
	server := iconServer(t, &hits)
	_, err := icon.Fetch(context.Background(), icon.Source{IconURL: server.URL + "/large.png"})
	if err == nil || !strings.Contains(err.Error(), "larger than") {
		t.Logf("hits = %d", hits.Load())
		t.Fatalf("oversized error = %v", err)
	}

	_, err = icon.Fetch(context.Background(), icon.Source{IconURL: server.URL + "/bad.png"})
	if err == nil || !strings.Contains(err.Error(), "supported bitmap") {
		t.Fatalf("unsupported error = %v", err)
	}
}

func TestFetchRejectsNonWebAddresses(t *testing.T) {
	if _, err := icon.Fetch(context.Background(), icon.Source{IconURL: "javascript:alert(1)"}); err == nil {
		t.Fatal("a javascript address should be refused")
	}
}
