package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/store"
)

func TestPostsListAggregatesFeeds(t *testing.T) {
	st := openStore(t)
	first, err := st.AddFeed(store.Feed{URL: "https://example.com/first", Title: "First"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := st.AddFeed(store.Feed{URL: "https://example.com/second", Title: "Second"})
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)
	for _, feed := range []store.Feed{first, second} {
		incoming := make([]store.Post, 0, 2)
		for i := range 2 {
			incoming = append(incoming, store.Post{
				GUID:        fmt.Sprintf("%d-%d", feed.ID, i),
				Title:       fmt.Sprintf("Post %d-%d", feed.ID, i),
				PublishedAt: base.Add(time.Duration(feed.ID)*time.Hour + time.Duration(i)*time.Minute),
			})
		}
		if _, err := st.MergePosts(feed.ID, incoming); err != nil {
			t.Fatal(err)
		}
	}

	page, err := NewPosts(st).List(PostQuery{FeedID: AggregateFeedID, Limit: 4})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 4 || len(page.Posts) != 4 {
		t.Fatalf("page = %+v", page)
	}
	if page.Unread != 4 {
		t.Errorf("unread = %d", page.Unread)
	}
	if page.Posts[0].GUID != "2-1" || page.Posts[2].GUID != "1-1" {
		t.Errorf("aggregate order = %+v", page.Posts)
	}

	ascending, err := NewPosts(st).List(PostQuery{
		FeedID:    AggregateFeedID,
		Limit:     4,
		SortOrder: SortOrderAscending,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ascending.Posts[0].GUID != "1-0" || ascending.Posts[3].GUID != "2-1" {
		t.Errorf("ascending order = %+v", ascending.Posts)
	}

	changed, err := NewPosts(st).MarkAllRead(AggregateFeedID)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 4 {
		t.Errorf("aggregate changed = %d", changed)
	}
}
