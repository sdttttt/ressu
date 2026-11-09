package status

import (
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/store"
)

func TestModelMirrorsStatusState(t *testing.T) {
	model := New("1.2.3")
	changed := make(chan struct{}, 4)
	model.SetChanged(func() { changed <- struct{}{} })

	fetched := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.UTC)
	feeds := []store.Feed{{ID: 1, Title: "A", LastFetchedAt: &fetched}}
	model.SetFeeds(FeedsSnapshot{Feeds: feeds, Unread: 3})
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetFeeds did not notify")
	}

	model.SetCurrentFeed(1)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetCurrentFeed did not notify")
	}

	snapshot := model.Snapshot()
	if snapshot.Version != "1.2.3" || snapshot.FeedCount != 1 || snapshot.Unread != 3 ||
		snapshot.Theme != store.ThemeSystem || snapshot.FontSize != store.FontSizeMedium {
		t.Fatalf("unexpected snapshot counts: %+v", snapshot)
	}
	if len(snapshot.Feeds) != 1 || snapshot.Feeds[0].ID != 1 || snapshot.CurrentFeedID != 1 {
		t.Fatalf("unexpected snapshot feeds: %+v", snapshot)
	}
	if snapshot.CurrentFeed == nil || snapshot.CurrentFeed.ID != 1 || snapshot.CurrentFeed.LastFetchedAt != &fetched {
		t.Fatalf("unexpected current feed: %+v", snapshot.CurrentFeed)
	}

	model.SetAppearance(store.ThemeSepia, store.FontSizeLarge)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetAppearance did not notify")
	}
	if snapshot := model.Snapshot(); snapshot.Theme != store.ThemeSepia || snapshot.FontSize != store.FontSizeLarge {
		t.Fatalf("appearance was not mirrored: %+v", snapshot)
	}

	model.SetCurrentFeed(1)
	select {
	case <-changed:
		t.Fatal("unchanged current feed notified")
	default:
	}
}
