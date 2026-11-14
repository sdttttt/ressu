package postheader

import (
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestHeaderShowsSelectedArticleAndTogglesStar(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
	model.SetFeeds(app.StatusFeedsSnapshot{Feeds: []app.Feed{{ID: 1, Title: "A"}}, Unread: 1})
	published := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.UTC)
	model.SetCurrentPost(
		app.PostSelection{FeedID: 1, GUID: "first"},
		app.PostSummary{
			FeedID:      1,
			GUID:        "first",
			Title:       "First article",
			URL:         "https://example.com/first",
			Author:      "Author",
			WordCount:   3,
			PublishedAt: published,
			Starred:     true,
		},
	)
	starred := make(chan bool, 1)
	snapshotRequested := make(chan bool, 1)
	header := New(model, Callbacks{
		ToggleStar: func(int64, string) (app.PostSummary, error) {
			model.SetPostUpdated(app.PostSummary{FeedID: 1, GUID: "first", Starred: false})
			starred <- false
			return app.PostSummary{FeedID: 1, GUID: "first", Starred: false}, nil
		},
		RequestSnapshot: func(*mygo.Window) {
			snapshotRequested <- true
		},
	})

	tester := ui.NewTester(header.View, 680, 96)
	tester.Frame()
	for _, text := range []string{
		"First article",
		"Author",
		"约 1 分钟 · 3 字",
		"A",
	} {
		if !tester.HasText(text) {
			t.Errorf("article header is missing %q", text)
		}
	}

	if err := tester.Click("取消星标"); err != nil {
		t.Fatalf("click starred action: %v", err)
	}
	select {
	case starredValue := <-starred:
		if starredValue {
			t.Fatal("star action did not toggle the article")
		}
	case <-time.After(time.Second):
		t.Fatal("clicking the starred action did not toggle the star")
	}
	tester.Frame()
	if !tester.HasText("加星标") {
		t.Error("article header did not refresh its star action")
	}

	if err := tester.Click("生成快照"); err != nil {
		t.Fatalf("click snapshot action: %v", err)
	}
	select {
	case <-snapshotRequested:
	case <-time.After(time.Second):
		t.Fatal("clicking the snapshot action did not request a snapshot")
	}
}
