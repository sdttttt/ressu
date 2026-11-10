package gui

import (
	"context"
	"testing"

	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

func TestFeedSidebarShowsSubscriptionActions(t *testing.T) {
	model := appstatus.New("")
	model.SetFeeds(appstatus.FeedsSnapshot{
		Feeds:  []store.Feed{{ID: 1, Title: "测试源", URL: "https://example.com/feed"}},
		Unread: 2,
	})
	sidebar := NewFeedSidebar(model, FeedSidebarCallbacks{
		Icon: func(context.Context, int64) (store.Icon, error) {
			return store.Icon{}, store.ErrNotFound
		},
	})

	tester := ui.NewTester(sidebar.View, 280, 520)
	tester.Frame()

	for _, text := range []string{"订阅源", "测试源"} {
		if !tester.HasText(text) {
			t.Errorf("sidebar is missing %q", text)
		}
	}
	for _, label := range []string{"浮动显示订阅源", "搜索订阅源", "添加订阅源", "刷新全部", "设置"} {
		if _, found := tester.Find(label); !found {
			t.Errorf("sidebar is missing %q action", label)
		}
	}
}
