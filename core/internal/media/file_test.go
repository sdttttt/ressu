package media

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileNameSuggestsSafeNames(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{"last path segment", "https://cdn.example.com/audio/ep1.mp3", "ep1.mp3"},
		{"query stripped", "https://cdn.example.com/dl.mp3?token=abc", "dl.mp3"},
		{"fragment stripped", "https://cdn.example.com/dl.mp3#part", "dl.mp3"},
		{"root path falls back", "https://cdn.example.com/", "attachment"},
		{"empty path falls back", "https://cdn.example.com", "attachment"},
		{"garbage falls back", "::not-a-url::", "attachment"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FileName(test.url); got != test.want {
				t.Errorf("FileName(%q) = %q, want %q", test.url, got, test.want)
			}
		})
	}
}

func TestFileNameTruncatesLongNames(t *testing.T) {
	long := "https://cdn.example.com/" + strings.Repeat("a", 200) + ".mp3"
	if got := FileName(long); len([]rune(got)) != maxFileNameRunes {
		t.Errorf("FileName length = %d, want %d", len([]rune(got)), maxFileNameRunes)
	}
}

// stubFileClient swaps the package-level download client for the test. The
// default policy refuses loopback and odd ports, so tests stub the transport
// the same way the image cache tests do.
func stubFileClient(t *testing.T, function roundTripFunc) {
	t.Helper()
	original := fileClient
	fileClient = &http.Client{Transport: function}
	t.Cleanup(func() { fileClient = original })
}

func TestDownloadFileStreamsToDisk(t *testing.T) {
	payload := strings.Repeat("episode-bytes", 4096)
	stubFileClient(t, func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Header:        http.Header{"Content-Type": []string{"audio/mpeg"}},
			Body:          io.NopCloser(strings.NewReader(payload)),
			ContentLength: int64(len(payload)),
			Request:       request,
		}, nil
	})

	dest := filepath.Join(t.TempDir(), "ep1.mp3")
	if err := DownloadFile(t.Context(), "https://cdn.example.com/ep1.mp3", dest); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != payload {
		t.Errorf("downloaded %d bytes, want %d", len(got), len(payload))
	}
}

func TestDownloadFileRejectsUnsafeAddresses(t *testing.T) {
	stubFileClient(t, func(request *http.Request) (*http.Response, error) {
		t.Error("the transport must not be reached for an unsafe address")
		return nil, errors.New("unexpected request")
	})
	dest := filepath.Join(t.TempDir(), "out")
	for _, source := range []string{
		"file:///etc/passwd",
		"ftp://example.com/file.mp3",
		"https://user:pass@example.com/file.mp3",
		"https://127.0.0.1/file.mp3",
		"https://example.com:8080/file.mp3",
	} {
		if err := DownloadFile(t.Context(), source, dest); err == nil {
			t.Errorf("DownloadFile accepted %q", source)
		}
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("a rejected download left a file behind")
	}
}

// failAfterReader delivers a few bytes and then errors like a reset
// connection, to exercise the partial-file cleanup.
type failAfterReader struct {
	remaining int
}

func (reader *failAfterReader) Read(buffer []byte) (int, error) {
	if reader.remaining <= 0 {
		return 0, errors.New("connection reset by test")
	}
	n := min(reader.remaining, len(buffer))
	reader.remaining -= n
	for i := range n {
		buffer[i] = 'x'
	}
	return n, nil
}

func TestDownloadFileRemovesPartialFileOnFailure(t *testing.T) {
	stubFileClient(t, func(request *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(&failAfterReader{remaining: 10}),
			Request:    request,
		}, nil
	})

	dest := filepath.Join(t.TempDir(), "out.mp3")
	if err := DownloadFile(t.Context(), "https://cdn.example.com/ep.mp3", dest); err == nil {
		t.Fatal("DownloadFile swallowed a mid-transfer failure")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Error("the failed download left a partial file behind")
	}
}
