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

func TestSidebarDragReordersFeeds(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{
			{ID: 1, Title: "源一", URL: "https://example.com/one"},
			{ID: 2, Title: "源二", URL: "https://example.com/two"},
			{ID: 3, Title: "源三", URL: "https://example.com/three"},
		},
	})
	type move struct {
		id int64
		to int
	}
	moves := make(chan move, 2)
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (app.Icon, error) {
			return app.Icon{}, app.ErrNotFound
		},
		Move: func(draggedID int64, to int) {
			moves <- move{draggedID, to}
		},
	})

	tester := ui.NewTester(sidebar.View, 280, 520)
	tester.Frame()
	one, ok := tester.Find("源一")
	if !ok {
		t.Fatal("the first feed row is missing")
	}
	// Dropping 源一 on the bottom half of 源三 places it after 源三: index
	// 2 in the resulting list.
	afterThree, ok := tester.Find("移动到「源三」之后")
	if !ok {
		t.Fatal("the after-drop zone of the third feed is missing")
	}
	startX, startY := one.X+one.W/2, one.Y+one.H/2
	targetX, targetY := afterThree.X+afterThree.W/2, afterThree.Y+afterThree.H/2

	tester.Press(startX, startY)
	for k := 1; k <= 4; k++ {
		f := float32(k) / 4
		tester.Move(startX+(targetX-startX)*f, startY+(targetY-startY)*f)
	}
	tester.Release(targetX, targetY)

	select {
	case got := <-moves:
		if got.id != 1 || got.to != 2 {
			t.Fatalf("move = %+v, want {1 2}", got)
		}
	case <-time.After(time.Second):
		t.Fatal("the drop did not report a move")
	}
}

func TestSidebarDisablesDragWhileFiltering(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{
			{ID: 1, Title: "源一", URL: "https://example.com/one"},
			{ID: 2, Title: "源二", URL: "https://example.com/two"},
		},
	})
	moved := make(chan struct{}, 1)
	sidebar := New(model, Callbacks{
		Icon: func(context.Context, int64) (app.Icon, error) {
			return app.Icon{}, app.ErrNotFound
		},
		Move: func(int64, int) { moved <- struct{}{} },
	})
	sidebar.search = "源一"

	tester := ui.NewTester(sidebar.View, 280, 520)
	tester.Frame()
	one, ok := tester.Find("源一")
	if !ok {
		t.Fatal("the filtered feed row is missing")
	}
	if _, found := tester.Find("移动到「源一」之前"); found {
		t.Error("drop zones should not render while filtering")
	}

	two, _ := tester.Find("源二")
	tester.Press(one.X+one.W/2, one.Y+one.H/2)
	tester.Move(two.X+two.W/2, two.Y+two.H/2)
	tester.Release(two.X+two.W/2, two.Y+two.H/2)

	select {
	case <-moved:
		t.Fatal("a move was reported while filtering")
	case <-time.After(100 * time.Millisecond):
	}
}
