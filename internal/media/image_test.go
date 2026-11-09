package media

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCacheGetDownloadsRemoteImage(t *testing.T) {
	cache := openTestCache(t)
	source := "https://cdn.example/viewer.png"
	data := testAVIF("avif", "mif1", "miaf")
	cache.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Header:        http.Header{"Content-Type": []string{avifContentType}},
			Body:          io.NopCloser(bytes.NewReader(data)),
			ContentLength: int64(len(data)),
			Request:       request,
		}, nil
	})}

	wantURL, ok := cache.LocalURL(source)
	if !ok {
		t.Fatal("source URL could not be mapped")
	}
	got, viewerURL, err := cache.Get(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) || viewerURL != wantURL {
		t.Fatalf("Get = %d bytes at %q; want %d bytes at %q", len(got), viewerURL, len(data), wantURL)
	}

	localSource := "http://" + Scheme + ".localhost/i/" + keyFor(source)
	got, viewerURL, err = cache.Get(context.Background(), localSource)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) || viewerURL != wantURL {
		t.Fatalf("local Get = %d bytes at %q; want %d bytes at %q", len(got), viewerURL, len(data), wantURL)
	}
}

func TestCacheGetServesStaleImageAfterFailedRefresh(t *testing.T) {
	cache := openTestCache(t)
	source := "https://cdn.example/stale.avif"
	data := testAVIF("avif", "mif1", "miaf")
	cache.client = &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Header:        http.Header{"Content-Type": []string{avifContentType}},
			Body:          io.NopCloser(bytes.NewReader(data)),
			ContentLength: int64(len(data)),
			Request:       request,
		}, nil
	})}

	wantURL, ok := cache.LocalURL(source)
	if !ok {
		t.Fatal("source URL could not be mapped")
	}
	if got, _, err := cache.Get(context.Background(), source); err != nil || !bytes.Equal(got, data) {
		t.Fatalf("initial Get = %d bytes, err = %v; want fresh image", len(got), err)
	}

	key := keyFor(source)
	cache.mu.Lock()
	entry := cache.entries[key]
	entry.FetchedAt = time.Now().Add(-freshTTL - time.Minute)
	entry.LastAttempt = entry.FetchedAt
	cache.entries[key] = entry
	cache.mu.Unlock()
	cache.client = &http.Client{Transport: failingTransport{}}

	got, viewerURL, err := cache.Get(context.Background(), source)
	if err != nil || !bytes.Equal(got, data) || viewerURL != wantURL {
		t.Fatalf("stale Get = %d bytes at %q, err = %v; want stale image", len(got), viewerURL, err)
	}
}

func TestCacheGetDecodesDataURL(t *testing.T) {
	cache := openTestCache(t)
	data := []byte("bitmap")
	source := "data:image/png;base64," + base64.StdEncoding.EncodeToString(data)

	got, viewerURL, err := cache.Get(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) || viewerURL != source {
		t.Fatalf("Get = %q at %q; want %q at its source", got, viewerURL, data)
	}
}

func TestCacheGetRejectsUnsafeSources(t *testing.T) {
	cache := openTestCache(t)
	for _, source := range []string{
		"file:///etc/passwd",
		"ressu-media://attacker.example/i/" + strings.Repeat("a", 64),
		"ressu-media://localhost/short",
		"data:image/png,aGVsbG8=",
	} {
		if _, _, err := cache.Get(context.Background(), source); err == nil {
			t.Fatalf("Get accepted %q", source)
		}
	}
}
