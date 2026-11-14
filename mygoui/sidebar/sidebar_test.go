package sidebar

import (
	"context"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestSidebarShowsSubscriptionActions(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 1, Title: "测试源", URL: "https://example.com/feed"}},
		Unread: 2,
	})
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (app.Icon, error) {
			return app.Icon{}, app.ErrNotFound
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
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
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
	model := app.NewStatusModel("")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 1, Title: "测试源", URL: "https://example.com/feed"}},
		Unread: 2,
	})
	selections := make(chan int64, 2)
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (app.Icon, error) {
			return app.Icon{}, app.ErrNotFound
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
	model := app.NewStatusModel("")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{
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

func TestSelectionPillSlidesBetweenFeeds(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{
			{ID: 1, Title: "源一", URL: "https://example.com/one"},
			{ID: 2, Title: "源二", URL: "https://example.com/two"},
		},
		Unread: 2,
	})
	model.SetCurrentFeed(1)
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (app.Icon, error) {
			return app.Icon{}, app.ErrNotFound
		},
	})

	tester := ui.NewTester(sidebar.View, 280, 520)
	tester.Frame()
	pillStart, ok := tester.Find("选中订阅源")
	if !ok {
		t.Fatal("the selection pill is missing")
	}
	rowStart, ok := tester.Find("源一")
	if !ok {
		t.Fatal("the first feed row is missing")
	}
	// The pill sits behind the selected row; the offset from the row text
	// is the same for every row of equal height.
	offset := pillStart.Y - rowStart.Y

	model.SetCurrentFeed(2)
	tester.Frame()
	pillMid, _ := tester.Find("选中订阅源")
	if pillMid.Y != pillStart.Y {
		t.Errorf("the pill jumped without time advancing: %v -> %v", pillStart.Y, pillMid.Y)
	}

	tester.Advance(2 * time.Second)
	tester.Frame()
	pillEnd, _ := tester.Find("选中订阅源")
	rowEnd, ok := tester.Find("源二")
	if !ok {
		t.Fatal("the second feed row is missing")
	}
	// Text baselines snap fractionally per row, so compare within a small
	// tolerance rather than exactly.
	if offsetEnd := pillEnd.Y - rowEnd.Y; offsetEnd < offset-0.01 || offsetEnd > offset+0.01 {
		t.Errorf("the pill did not slide to the second row: offset %v, want %v", offsetEnd, offset)
	}
}

func TestSelectionPillJumpsWhenReduceMotion(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetReduceMotion(true)
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{
			{ID: 1, Title: "源一", URL: "https://example.com/one"},
			{ID: 2, Title: "源二", URL: "https://example.com/two"},
		},
		Unread: 2,
	})
	model.SetCurrentFeed(1)
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (app.Icon, error) {
			return app.Icon{}, app.ErrNotFound
		},
	})

	tester := ui.NewTester(sidebar.View, 280, 520)
	tester.Frame()
	pillStart, ok := tester.Find("选中订阅源")
	if !ok {
		t.Fatal("the selection pill is missing")
	}
	rowStart, ok := tester.Find("源一")
	if !ok {
		t.Fatal("the first feed row is missing")
	}
	offset := pillStart.Y - rowStart.Y

	model.SetCurrentFeed(2)
	tester.Frame()
	pillEnd, _ := tester.Find("选中订阅源")
	rowEnd, ok := tester.Find("源二")
	if !ok {
		t.Fatal("the second feed row is missing")
	}
	if offsetEnd := pillEnd.Y - rowEnd.Y; offsetEnd < offset-0.01 || offsetEnd > offset+0.01 {
		t.Errorf("the pill did not jump to the second row: offset %v, want %v", offsetEnd, offset)
	}
}
