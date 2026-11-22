package service

import (
	"errors"
	"testing"

	"github.com/sdttttt/ressu/core/internal/store"
)

func TestMoveReordersAndBroadcasts(t *testing.T) {
	st := openStore(t)
	var ids []int64
	for _, url := range []string{"https://a.example/feed", "https://b.example/feed", "https://c.example/feed"} {
		feed, err := st.AddFeed(store.Feed{URL: url})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, feed.ID)
	}

	feeds := NewFeeds(st)
	if err := feeds.Move(ids[2], 0); err != nil {
		t.Fatal(err)
	}
	list := feeds.List().Feeds
	if len(list) != 3 || list[0].ID != ids[2] || list[1].ID != ids[0] || list[2].ID != ids[1] {
		t.Fatalf("order after move = %+v", list)
	}

	if err := feeds.Move(9999, 0); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("Move unknown id error = %v", err)
	}
}
