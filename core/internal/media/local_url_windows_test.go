//go:build windows

package media

import (
	"strings"
	"testing"
)

func TestLocalURLUsesWebView2SchemeHost(t *testing.T) {
	key := keyFor("https://cdn.example/photo.png")
	got := localURL(key)

	want := "http://" + Scheme + ".localhost/i/" + key
	if got != want || !strings.HasPrefix(got, "http://ressu-media.localhost/") {
		t.Fatalf("localURL = %q, want %q", got, want)
	}
}
