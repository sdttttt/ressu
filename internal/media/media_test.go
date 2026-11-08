package media

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRewriteHTMLPreparesRemoteImages(t *testing.T) {
	cache := openTestCache(t)
	source := "https://cdn.example/cover.png"
	output := cache.RewriteHTML(
		`<p>before</p><img src="`+source+`" srcset="small.png 1x" loading="lazy">`,
		"https://example.com/post",
	)

	key := keyFor(source)
	if !strings.Contains(output, Scheme+"://localhost/i/"+key) ||
		strings.Contains(output, "cdn.example") || strings.Contains(output, "srcset") {
		t.Fatalf("unexpected rewritten HTML: %s", output)
	}
	entry, cached := cache.entries[key]
	if !cached || entry.SourceURL != "https://cdn.example/cover.png" || entry.Size != 0 {
		t.Fatalf("prepared entry = %+v, cached = %t", entry, cached)
	}
}

func TestRewriteHTMLKeepsDataImagesAndDropsUnsafeImages(t *testing.T) {
	cache := openTestCache(t)
	input := `<img src="data:image/png;base64,aGVsbG8=">` +
		`<img src="http://127.0.0.1:8080/private.png">` +
		`<img src="https://cdn.example/real.png">`
	output := cache.RewriteHTML(input, "")

	if !strings.Contains(output, `src="data:image/png;base64,aGVsbG8="`) {
		t.Fatalf("data image was removed: %s", output)
	}
	if strings.Contains(output, "127.0.0.1") {
		t.Fatalf("private image stayed reachable: %s", output)
	}
	if !strings.Contains(output, Scheme+"://localhost/i/") {
		t.Fatalf("public image was not rewritten: %s", output)
	}
}

func TestHandlerServesCachedImage(t *testing.T) {
	cache := openTestCache(t)
	source := "https://cdn.example/photo.png"
	key := keyFor(source)
	data := []byte("PNG image bytes")
	entry := Entry{
		Key:         key,
		SourceURL:   source,
		ContentType: "image/png",
		FetchedAt:   time.Now(),
		LastAttempt: time.Now(),
		Size:        int64(len(data)),
	}
	if err := writeData(cache.dataFile(key), data); err != nil {
		t.Fatal(err)
	}
	if err := saveEntry(cache.dir, entry); err != nil {
		t.Fatal(err)
	}
	cache.mu.Lock()
	cache.entries[key] = entry
	cache.total += entry.Size
	cache.mu.Unlock()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, Scheme+"://localhost/i/"+key, nil)
	cache.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || !bytes.Equal(recorder.Body.Bytes(), data) ||
		recorder.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("handler answered %d: %q %v", recorder.Code, recorder.Body.String(), recorder.Header())
	}
}

func TestHandlerRejectsUnknownKeysAndMethods(t *testing.T) {
	cache := openTestCache(t)
	requests := []*http.Request{
		httptest.NewRequest(http.MethodPost, Scheme+"://localhost/i/"+strings.Repeat("a", 64), nil),
		httptest.NewRequest(http.MethodGet, Scheme+"://localhost/i/short", nil),
		httptest.NewRequest(http.MethodGet, Scheme+"://example.com/i/"+strings.Repeat("a", 64), nil),
		httptest.NewRequest(http.MethodGet, Scheme+"://localhost/i/"+keyFor("https://cdn.example/missing.png"), nil),
	}
	for _, request := range requests {
		recorder := httptest.NewRecorder()
		cache.ServeHTTP(recorder, request)
		if recorder.Code == http.StatusOK {
			t.Fatalf("%s unexpectedly succeeded", request.URL)
		}
	}
}

func TestGetServesStaleImageAfterFailure(t *testing.T) {
	cache := openTestCache(t)
	source := "https://cdn.example/stale.png"
	key := keyFor(source)
	data := []byte("stale PNG")
	entry := Entry{
		Key:         key,
		SourceURL:   source,
		ContentType: "image/png",
		FetchedAt:   time.Now().Add(-freshTTL - time.Minute),
		LastAttempt: time.Now().Add(-freshTTL - time.Minute),
		Size:        int64(len(data)),
	}
	if err := writeData(cache.dataFile(key), data); err != nil {
		t.Fatal(err)
	}
	if err := saveEntry(cache.dir, entry); err != nil {
		t.Fatal(err)
	}
	cache.mu.Lock()
	cache.entries[key] = entry
	cache.total += entry.Size
	cache.mu.Unlock()
	cache.client = &http.Client{Transport: failingTransport{}}

	fresh, err := cache.get(context.Background(), key)
	if err == nil || fresh.Size != entry.Size || fresh.FailedAt.IsZero() {
		t.Fatalf("get = %+v, err = %v; want stale bytes and a failed refresh", fresh, err)
	}
	if _, err := os.Stat(cache.dataFile(key)); err != nil {
		t.Fatalf("stale data was removed: %v", err)
	}
}

func TestSourceValidationRejectsUnsafeAddresses(t *testing.T) {
	for _, source := range []string{
		"file:///etc/passwd",
		"javascript:alert(1)",
		"http://user:password@example.com/image.png",
		"http://example.com:8080/image.png",
		"http://[fd00::1]/image.png",
	} {
		if _, ok := normalizeSource(source); ok {
			t.Fatalf("normalizeSource accepted %q", source)
		}
		if validateSource(source) == nil {
			t.Fatalf("validateSource accepted %q", source)
		}
	}
}

func TestPublicAddressRejectsPrivateNetworks(t *testing.T) {
	for _, source := range []string{
		"127.0.0.1",
		"10.0.0.1",
		"169.254.169.254",
		"100.64.0.1",
		"::1",
		"fe80::1",
	} {
		if publicAddress(net.ParseIP(source)) {
			t.Fatalf("publicAddress accepted %q", source)
		}
	}
	if !publicAddress(net.ParseIP("93.184.216.34")) {
		t.Fatal("publicAddress rejected a public address")
	}
}

type failingTransport struct{}

func (failingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("test transport failure")
}

func openTestCache(t *testing.T) *Cache {
	t.Helper()
	cache, err := Open(filepath.Join(t.TempDir(), "media"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cache.Close() })
	return cache
}
