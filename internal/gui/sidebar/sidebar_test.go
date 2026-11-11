package sidebar

import (
	"context"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

func TestSidebarShowsSubscriptionActions(t *testing.T) {
	model := appstatus.New("")
	model.SetFeeds(appstatus.FeedsSnapshot{
		Feeds:  []store.Feed{{ID: 1, Title: "测试源", URL: "https://example.com/feed"}},
		Unread: 2,
	})
	sidebar := New(model, Callbacks{
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

func TestSidebarMatchesWebviewTextSize(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
	sidebar := New(model, Callbacks{})

	var fontSize float32
	tester := ui.NewTester(func(c *ui.Context) {
		fontSize = sidebar.Theme(c).FontSize
	}, 280, 520)
	tester.Frame()

	if fontSize != textSize {
		t.Errorf("sidebar font size = %v, want %v", fontSize, textSize)
	}
}

func TestSidebarClicksToggleSelectedFeed(t *testing.T) {
	model := appstatus.New("")
	model.SetFeeds(appstatus.FeedsSnapshot{
		Feeds:  []store.Feed{{ID: 1, Title: "测试源", URL: "https://example.com/feed"}},
		Unread: 2,
	})
	selections := make(chan int64, 2)
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (store.Icon, error) {
			return store.Icon{}, store.ErrNotFound
		},
		Select: func(feedID int64) {
			selections <- feedID
		},
	})

	tester := ui.NewTester(sidebar.View, 280, 520)
	tester.Frame()
	if err := tester.Click("测试源"); err != nil {
		t.Fatalf("click feed: %v", err)
	}
	select {
	case feedID := <-selections:
		if feedID != 1 {
			t.Fatalf("first selection = %d; want 1", feedID)
		}
		model.SetCurrentFeed(feedID)
	case <-time.After(time.Second):
		t.Fatal("first selection was not reported")
	}
	tester.Frame()

	if err := tester.Click("测试源"); err != nil {
		t.Fatalf("click selected feed: %v", err)
	}
	select {
	case feedID := <-selections:
		if feedID != aggregateFeedID {
			t.Fatalf("second selection = %d; want %d", feedID, aggregateFeedID)
		}
	case <-time.After(time.Second):
		t.Fatal("second selection was not reported")
	}
}

func TestSidebarSelectNeighbourClampsToFeedBounds(t *testing.T) {
	model := appstatus.New("")
	model.SetFeeds(appstatus.FeedsSnapshot{
		Feeds: []store.Feed{
			{ID: 1, Title: "第一个", URL: "https://example.com/one"},
			{ID: 2, Title: "第二个", URL: "https://example.com/two"},
			{ID: 3, Title: "第三个", URL: "https://example.com/three"},
		},
	})
	selections := make(chan int64, 4)
	sidebar := New(model, Callbacks{
		Select: func(feedID int64) {
			selections <- feedID
		},
	})
	model.SetCurrentFeed(2)

	sidebar.SelectNeighbour(-1)
	assertSelection(t, selections, 1)
	model.SetCurrentFeed(1)

	sidebar.SelectNeighbour(-1)
	assertSelection(t, selections, 1)

	sidebar.SelectNeighbour(1)
	assertSelection(t, selections, 2)
}

func assertSelection(t *testing.T, selections <-chan int64, want int64) {
	t.Helper()
	select {
	case got := <-selections:
		if got != want {
			t.Fatalf("selection = %d; want %d", got, want)
		}
	case <-time.After(time.Second):
		t.Fatal("selection was not reported")
	}
}
