package addfeed

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

// stubCallbacks returns inert callbacks for rendering tests.
func stubCallbacks(prefs app.Prefs) Callbacks {
	return Callbacks{
		Appearance: func() app.Prefs { return prefs },
		Test: func(context.Context, string, int64) (Preview, error) {
			return Preview{}, nil
		},
		Add:        func(context.Context, string, string, string, int) (app.Feed, error) { return app.Feed{}, nil },
		Save:       func(context.Context, int64, string, string, string, int) (app.Feed, error) { return app.Feed{}, nil },
		ImportOPML: func(*mygo.Window) (Summary, error) { return Summary{}, nil },
	}
}

// timePtr returns a pointer to value, for feed timestamps.
func timePtr(value time.Time) *time.Time { return &value }

func TestAddFeedViewRendersDraft(t *testing.T) {
	prefs := app.DefaultPrefs()
	view := &addFeedView{
		callbacks: stubCallbacks(prefs),
		prefs:     prefs,
		url:       "https://example.com/feed.xml",
		testedURL: "https://example.com/feed.xml",
		title:     "示例",
		preview: &Preview{
			Title:       "示例订阅源",
			SiteURL:     "https://example.com/",
			Description: "一个测试订阅源",
			Language:    "zh-CN",
		},
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	for _, want := range []string{
		"订阅地址",
		"订阅名称",
		"示例",
		"https://example.com/",
		"一个测试订阅源",
		"语言：zh-CN",
		"测试地址",
		"添加订阅",
		"导入 OPML…",
		"远程资源访问方式",
	} {
		if !tester.HasText(want) {
			t.Errorf("Add-feed window is missing %q", want)
		}
	}
	if tester.HasText("文章保留数") {
		t.Error("Add-feed window should not render the per-feed keep count")
	}
	if tester.HasText("状态") {
		t.Error("Add-feed window should not render the feed health section")
	}
}

func TestAddFeedViewRendersError(t *testing.T) {
	prefs := app.DefaultPrefs()
	view := &addFeedView{
		callbacks: stubCallbacks(prefs),
		prefs:     prefs,
		url:       "https://example.com/feed.xml",
		err:       "操作失败：" + errors.New("无法读取订阅源").Error(),
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	if !tester.HasText("操作失败：无法读取订阅源") {
		t.Error("Add-feed window is missing the operation error")
	}
}

func TestAddFeedViewRendersEditMode(t *testing.T) {
	prefs := app.DefaultPrefs()
	view := &addFeedView{
		callbacks:       stubCallbacks(prefs),
		prefs:           prefs,
		edit:            &app.Feed{ID: 7, URL: "https://example.com/feed.xml", Title: "自定义名称", CustomTitle: true},
		url:             "https://example.com/feed.xml",
		title:           "自定义名称",
		remoteResources: app.ResourceAccessGoProxy,
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	for _, want := range []string{"订阅地址", "订阅名称", "保存修改", "测试地址", "远程资源访问方式", "文章保留数"} {
		if !tester.HasText(want) {
			t.Errorf("Edit-feed window is missing %q", want)
		}
	}
	for _, unwanted := range []string{"添加订阅", "导入 OPML…"} {
		if tester.HasText(unwanted) {
			t.Errorf("Edit-feed window should not render %q", unwanted)
		}
	}
}

func TestAddFeedViewRendersFeedHealth(t *testing.T) {
	prefs := app.DefaultPrefs()
	openedAt := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	view := &addFeedView{
		callbacks: stubCallbacks(prefs),
		prefs:     prefs,
		edit: &app.Feed{
			ID:            7,
			URL:           "https://example.com/feed.xml",
			LastFetchedAt: timePtr(openedAt.Add(-2 * time.Hour)),
		},
		openedAt: openedAt,
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	for _, want := range []string{"状态", "正常", "上次刷新", "2 小时前"} {
		if !tester.HasText(want) {
			t.Errorf("Edit-feed window is missing %q", want)
		}
	}
	for _, unwanted := range []string{"错误信息", "上次刷新失败", "尚未刷新"} {
		if tester.HasText(unwanted) {
			t.Errorf("Edit-feed window should not render %q", unwanted)
		}
	}
}

func TestAddFeedViewRendersFeedFailure(t *testing.T) {
	prefs := app.DefaultPrefs()
	openedAt := time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC)
	view := &addFeedView{
		callbacks: stubCallbacks(prefs),
		prefs:     prefs,
		edit: &app.Feed{
			ID:            7,
			URL:           "https://example.com/feed.xml",
			LastFetchedAt: timePtr(openedAt.Add(-30 * time.Minute)),
			LastError:     "the server answered 500 Forbidden",
		},
		openedAt: openedAt,
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	for _, want := range []string{"上次刷新失败", "错误信息", "the server answered 500 Forbidden", "上次刷新", "30 分钟前"} {
		if !tester.HasText(want) {
			t.Errorf("Edit-feed window is missing %q", want)
		}
	}
	if tester.HasText("正常") {
		t.Error("Edit-feed window should not report a failed feed as healthy")
	}
}

func TestAddFeedViewRendersUnfetchedFeed(t *testing.T) {
	prefs := app.DefaultPrefs()
	view := &addFeedView{
		callbacks: stubCallbacks(prefs),
		prefs:     prefs,
		edit:      &app.Feed{ID: 7, URL: "https://example.com/feed.xml"},
		openedAt:  time.Date(2026, 11, 28, 12, 0, 0, 0, time.UTC),
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	for _, want := range []string{"状态", "尚未刷新"} {
		if !tester.HasText(want) {
			t.Errorf("Edit-feed window is missing %q", want)
		}
	}
	for _, unwanted := range []string{"上次刷新", "错误信息", "正常"} {
		if tester.HasText(unwanted) {
			t.Errorf("Edit-feed window should not render %q", unwanted)
		}
	}
}
