package mygoui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestStatusBarVersionOpensAbout(t *testing.T) {
	model := app.NewStatusModel("26.11.8")
	opened := false
	status := NewStatusBar(model, func() {
		opened = true
	})

	tester := ui.NewTester(status.View, 640, 60)
	if err := tester.Click("v26.11.8"); err != nil {
		t.Fatal(err)
	}
	if !opened {
		t.Fatal("clicking the version did not open the About window")
	}
}

func TestStatusBarShowsFeedErrorBesideLastRefresh(t *testing.T) {
	model := app.NewStatusModel("26.11.8")
	fetched := time.Now().Add(-2 * time.Hour)
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{{
			ID:            1,
			Title:         "测试源",
			URL:           "https://example.com/feed",
			LastFetchedAt: &fetched,
			LastError:     "the server answered 404 Not Found",
		}},
	})
	model.SetCurrentFeed(1)
	status := NewStatusBar(model, nil)

	tester := ui.NewTester(status.View, 640, 60)
	tester.Frame()

	for _, text := range []string{"刷新失败:the server answered 404 Not Found", "上次刷新 2 小时前"} {
		if !tester.HasText(text) {
			t.Errorf("status bar is missing %q", text)
		}
	}
}
