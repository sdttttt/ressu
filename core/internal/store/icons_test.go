package store

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func TestIconCache(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}

	fetchedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	icon := Icon{
		Meta: IconMeta{
			SourceURL:   "https://a.example/icon.png",
			ContentType: "image/png",
			FetchedAt:   fetchedAt,
		},
		Data: []byte{1, 2, 3},
	}
	if err := s.PutIcon(feed.ID, icon); err != nil {
		t.Fatal(err)
	}

	got, err := s.Icon(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Meta != icon.Meta || !bytes.Equal(got.Data, icon.Data) {
		t.Errorf("icon = %+v, want %+v", got, icon)
	}

	icon.Data = []byte{4, 5}
	if err := s.PutIcon(feed.ID, icon); err != nil {
		t.Fatal(err)
	}
	got, err = s.Icon(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Data, icon.Data) {
		t.Errorf("replacement data = %v, want %v", got.Data, icon.Data)
	}

	if err := s.DeleteIcon(feed.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Icon(feed.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("deleted icon error = %v", err)
	}
	if err := s.PutIcon(feed.ID, icon); err != nil {
		t.Fatal(err)
	}

	if err := s.RemoveFeed(feed.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Icon(feed.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("removed icon error = %v", err)
	}
}

func TestPutIconRejectsOversizedData(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}

	icon := Icon{Data: make([]byte, maxIconBytes+1)}
	if err := s.PutIcon(feed.ID, icon); err == nil {
		t.Fatal("an oversized icon should be refused")
	}
}
