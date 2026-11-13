package status

import (
	"testing"
	"time"

	"github.com/sdttttt/ressu/core/store"
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

	model.SetLayout(true, 27)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetLayout did not notify")
	}
	layout := model.Snapshot()
	if !layout.SidebarCollapsed || layout.SidebarPercent != 27 {
		t.Fatalf("layout was not mirrored: %+v", layout)
	}

	model.SetLayout(false, 101)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetLayout did not notify")
	}
	if layout := model.Snapshot(); layout.SidebarCollapsed || layout.SidebarPercent != 18 {
		t.Fatalf("out-of-range sidebar percent was not clamped: %+v", layout)
	}

	selected := store.PostSummary{FeedID: 1, GUID: "first", Title: "First", Read: false}
	model.SetCurrentPost(PostSelection{FeedID: 1, GUID: "first"}, selected)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetCurrentPost did not notify")
	}
	if selected := model.Snapshot().SelectedPost; selected.GUID != "first" || selected.Read {
		t.Fatalf("article header summary was not mirrored: %+v", selected)
	}

	model.SetPostUpdated(store.PostSummary{FeedID: 1, GUID: "first", Title: "First", Read: true})
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetPostUpdated did not notify")
	}
	if selected := model.Snapshot().SelectedPost; !selected.Read {
		t.Fatalf("post update did not refresh the selected summary: %+v", selected)
	}

	model.SetCurrentFeed(1)
	select {
	case <-changed:
		t.Fatal("unchanged current feed notified")
	default:
	}
	if current := model.Snapshot().CurrentPost; current.GUID != "first" {
		t.Fatalf("unchanged feed cleared the article selection: %+v", current)
	}

	model.SetCurrentFeed(2)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("changing the feed did not notify")
	}
	if current := model.Snapshot().CurrentPost; current.GUID != "" {
		t.Fatalf("changing the feed kept the article selection: %+v", current)
	}
	if selected := model.Snapshot().SelectedPost; selected.GUID != "" {
		t.Fatalf("changing the feed kept the selected summary: %+v", selected)
	}
}

func TestModelMirrorsReduceMotion(t *testing.T) {
	model := New("")
	if model.Snapshot().ReduceMotion {
		t.Fatal("reduce motion is on by default")
	}

	changed := make(chan struct{}, 2)
	model.SetChanged(func() { changed <- struct{}{} })

	model.SetReduceMotion(true)
	select {
	case <-changed:
	case <-time.After(time.Second):
		t.Fatal("SetReduceMotion did not notify")
	}
	if !model.Snapshot().ReduceMotion {
		t.Fatal("reduce motion was not mirrored")
	}

	model.SetReduceMotion(true)
	select {
	case <-changed:
		t.Fatal("unchanged reduce motion notified")
	default:
	}
}
