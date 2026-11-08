package media

import (
	"bytes"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandlerLogsAndDeduplicatesUnavailableImages(t *testing.T) {
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	cache, err := Open(filepath.Join(t.TempDir(), "media"))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()

	key := keyFor("https://cdn.example/missing.png")
	request := httptest.NewRequest(http.MethodGet, Scheme+"://localhost/i/"+key, nil)
	recorder := httptest.NewRecorder()
	cache.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d; want unavailable", recorder.Code)
	}

	recorder = httptest.NewRecorder()
	cache.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("repeat status = %d; want unavailable", recorder.Code)
	}

	logs := output.String()
	if strings.Count(logs, "media image unavailable") != 1 {
		t.Fatalf("expected one deduplicated log, got: %s", logs)
	}
	if strings.Contains(logs, "cdn.example") {
		t.Fatalf("log leaked the source URL: %s", logs)
	}
}

func TestDiagnosticErrorRedactsURLs(t *testing.T) {
	cause := errors.New(`Get "https://user:secret@cdn.example/private.png": dial tcp: failure`)
	logged := diagnosticError(cause)

	if !strings.Contains(logged, "https://[redacted]") {
		t.Fatalf("expected redacted URL, got: %s", logged)
	}
	if strings.Contains(logged, "cdn.example") || strings.Contains(logged, "secret") {
		t.Fatalf("diagnostic error leaked the source URL: %s", logged)
	}
}
