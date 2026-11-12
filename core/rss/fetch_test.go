package rss

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sdttttt/ressu/core/limitreader"
)

func TestFetchParsesFeedFormats(t *testing.T) {
	cases := []struct {
		name    string
		fixture string
		title   string
	}{
		{"rss", "rss2.xml", "Example Feed"},
		{"atom", "atom.xml", "Atom Example"},
		{"json", "jsonfeed.json", "JSON Example"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := []byte(readFixture(t, tc.fixture))
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(data)
			}))
			t.Cleanup(server.Close)

			result, err := fetch(context.Background(), server.URL, "", "", int64(len(data)))
			if err != nil {
				t.Fatal(err)
			}
			if result.Feed.Meta.Title != tc.title {
				t.Errorf("title = %q, want %q", result.Feed.Meta.Title, tc.title)
			}
			if len(result.Feed.Items) == 0 {
				t.Error("parsed feed has no items")
			}
		})
	}
}

func TestFetchRejectsOversizedBody(t *testing.T) {
	data := []byte(readFixture(t, "rss2.xml"))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)

	_, err := fetch(context.Background(), server.URL, "", "", int64(len(data)-1))
	if !errors.Is(err, limitreader.ErrTooLarge) {
		t.Fatalf("error = %v, want limitreader.ErrTooLarge", err)
	}
}

func TestFetchRejectsTrailingDataOverLimit(t *testing.T) {
	feed := []byte(readFixture(t, "rss2.xml"))
	body := append(append([]byte(nil), feed...), strings.Repeat(" ", 20_000)...)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	_, err := fetch(context.Background(), server.URL, "", "", int64(len(feed)+8_192))
	if !errors.Is(err, limitreader.ErrTooLarge) {
		t.Fatalf("error = %v, want limitreader.ErrTooLarge", err)
	}
}
