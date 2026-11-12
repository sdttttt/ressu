package gui

import (
	"context"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/core/status"
	"github.com/sdttttt/ressu/core/store"
	appdialog "github.com/sdttttt/ressu/internal/gui/dialog"
	appsidebar "github.com/sdttttt/ressu/internal/gui/sidebar"
)

func TestAppViewRendersNativeHelpAndFeedShortcuts(t *testing.T) {
	model := appstatus.New("")
	model.SetFeeds(appstatus.FeedsSnapshot{
		Feeds: []store.Feed{
			{ID: 1, Title: "第一个", URL: "https://example.com/one"},
			{ID: 2, Title: "第二个", URL: "https://example.com/two"},
		},
	})
	selections := make(chan int64, 2)
	sidebar := appsidebar.New(model, appsidebar.Callbacks{
		Icon: func(context.Context, int64) (store.Icon, error) {
			return store.Icon{}, store.ErrNotFound
		},
		Select: func(feedID int64) {
			selections <- feedID
		},
	})
	help := appdialog.NewHelp(func() (string, string) {
		return "system", "small"
	})
	view := NewAppView(model, func() {}, nil, nil, help, sidebar, nil, nil)
	model.SetCurrentFeed(2)

	tester := ui.NewTester(view.View, 900, 600)
	tester.Frame()
	tester.Key(ui.Shift, ui.KeySlash)
	tester.Frame()
	if !tester.HasText("快捷键") || !tester.HasText("上一个订阅源") {
		t.Fatal("help shortcut did not show the native help modal")
	}

	tester.Key(0, ui.KeyEscape)
	tester.Frame()
	tester.Key(0, ui.KeyL)
	select {
	case feedID := <-selections:
		if feedID != 2 {
			t.Fatalf("next feed selection = %d; want 2", feedID)
		}
	case <-time.After(time.Second):
		t.Fatal("next feed selection was not reported")
	}
}
