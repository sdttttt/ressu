package gui

import (
	"context"
	"errors"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/store"
)

func TestAddFeedViewRendersDraft(t *testing.T) {
	prefs := store.DefaultPrefs()
	view := &addFeedView{
		callbacks: AddFeedCallbacks{
			Appearance: func() store.Prefs { return prefs },
			Test: func(context.Context, string) (FeedPreview, error) {
				return FeedPreview{}, nil
			},
			Add: func(context.Context, string, string) (store.Feed, error) {
				return store.Feed{}, nil
			},
			ImportOPML: func(*Window) (ImportSummary, error) { return ImportSummary{}, nil },
		},
		prefs:     prefs,
		url:       "https://example.com/feed.xml",
		testedURL: "https://example.com/feed.xml",
		title:     "示例",
		preview: &FeedPreview{
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
	} {
		if !tester.HasText(want) {
			t.Errorf("Add-feed window is missing %q", want)
		}
	}
}

func TestAddFeedViewRendersError(t *testing.T) {
	prefs := store.DefaultPrefs()
	view := &addFeedView{
		callbacks: AddFeedCallbacks{
			Appearance: func() store.Prefs { return prefs },
			Test: func(context.Context, string) (FeedPreview, error) {
				return FeedPreview{}, nil
			},
			Add: func(context.Context, string, string) (store.Feed, error) {
				return store.Feed{}, nil
			},
			ImportOPML: func(*Window) (ImportSummary, error) { return ImportSummary{}, nil },
		},
		prefs: prefs,
		url:   "https://example.com/feed.xml",
		err:   "操作失败：" + errors.New("无法读取订阅源").Error(),
	}

	tester := ui.NewTester(view.draw, addFeedWidth, addFeedHeight)
	if !tester.HasText("操作失败：无法读取订阅源") {
		t.Error("Add-feed window is missing the operation error")
	}
}
