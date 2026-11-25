package postlist

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/png"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestPostListShowsArticles(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{Unread: 1})
	loaded := make(chan struct{}, 1)
	selected := make(chan app.PostSummary, 1)
	directions := make(chan int, 1)
	now := time.Now()
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			defer func() {
				loaded <- struct{}{}
			}()
			return app.PostPage{
				Posts: []app.PostSummary{
					{
						FeedID:      1,
						GUID:        "first",
						Title:       "First article",
						Summary:     "First summary",
						WordCount:   1234,
						PublishedAt: now,
					},
					{FeedID: 1, GUID: "second", Title: "Second article", Summary: "Second summary", Read: true},
				},
				Total:  2,
				Unread: 1,
			}, nil
		},
		Select: func(post app.PostSummary, direction int) {
			directions <- direction
			selected <- post
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	select {
	case <-loaded:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	tester.Frame()

	for _, text := range []string{
		"聚合源",
		"2 篇",
		"1 未读",
		"First article",
		"Second article",
		"刚刚",
		"约 5 分钟 · 1,234 字",
	} {
		if !tester.HasText(text) {
			t.Errorf("article list is missing %q", text)
		}
	}

	if err := tester.Click("First article"); err != nil {
		t.Fatalf("click first article: %v", err)
	}
	select {
	case post := <-selected:
		if post.FeedID != 1 || post.GUID != "first" || post.Title != "First article" {
			t.Errorf("selected post = %+v, want feed 1 guid first", post)
		}
	case <-time.After(time.Second):
		t.Fatal("clicking an article did not select it")
	}
	if direction := <-directions; direction != 0 {
		t.Errorf("click selection direction = %d, want 0", direction)
	}
}

func TestPostListShowsFeedNameOnlyInAggregateView(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds: []app.Feed{
			{ID: 1, Title: "科技源播报"},
			{ID: 2, Title: "设计源周报"},
		},
		Unread: 2,
	})
	loaded := make(chan struct{}, 2)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			defer func() { loaded <- struct{}{} }()
			if query.FeedID == 1 {
				return app.PostPage{
					Posts:  []app.PostSummary{{FeedID: 1, GUID: "a", Title: "First article"}},
					Total:  1,
					Unread: 1,
				}, nil
			}
			return app.PostPage{
				Posts: []app.PostSummary{
					{FeedID: 1, GUID: "a", Title: "First article"},
					{FeedID: 2, GUID: "b", Title: "Second article"},
				},
				Total:  2,
				Unread: 2,
			}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	select {
	case <-loaded:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	tester.Frame()

	for _, text := range []string{"科技源播报", "设计源周报", "·"} {
		if !tester.HasText(text) {
			t.Errorf("aggregate article list is missing %q", text)
		}
	}

	model.SetCurrentFeed(1)
	tester.Frame()
	select {
	case <-loaded:
	case <-time.After(time.Second):
		t.Fatal("article list did not reload the selected feed")
	}
	tester.Frame()

	if tester.HasText("设计源周报") {
		t.Error("single-feed view shows another feed's name")
	}
}

func TestPostListNavigationShortcuts(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{Unread: 2})
	loaded := make(chan struct{}, 1)
	selected := make(chan app.PostSummary, 8)
	directions := make(chan int, 8)
	starred := make(chan string, 1)
	postList := New(model, Callbacks{
		List: func(app.PostQuery) (app.PostPage, error) {
			defer func() { loaded <- struct{}{} }()
			return app.PostPage{
				Posts: []app.PostSummary{
					{FeedID: 1, GUID: "first", Title: "First article"},
					{FeedID: 1, GUID: "second", Title: "Second article", Read: true},
					{FeedID: 1, GUID: "third", Title: "Third article"},
				},
				Total:  3,
				Unread: 2,
			}, nil
		},
		Select: func(post app.PostSummary, direction int) {
			selection := app.PostSelection{FeedID: post.FeedID, GUID: post.GUID}
			model.SetCurrentPost(selection, post)
			directions <- direction
			selected <- post
		},
		ToggleStar: func(_ int64, guid string) (app.PostSummary, error) {
			starred <- guid
			return app.PostSummary{}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	select {
	case <-loaded:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	tester.Frame()

	expectSelection := func(guid string, wantDirection int) {
		t.Helper()
		select {
		case post := <-selected:
			if post.GUID != guid {
				t.Errorf("selected guid = %q, want %q", post.GUID, guid)
			}
		case <-time.After(time.Second):
			t.Fatalf("pressing a shortcut did not select %q", guid)
		}
		select {
		case direction := <-directions:
			if direction != wantDirection {
				t.Errorf("selection direction = %d, want %d", direction, wantDirection)
			}
		case <-time.After(time.Second):
			t.Fatalf("selection of %q did not report a direction", guid)
		}
	}

	tester.Key(0, ui.KeyJ)
	expectSelection("first", 0)
	tester.Key(0, ui.KeyJ)
	expectSelection("second", 1)
	tester.Key(0, ui.KeyN)
	expectSelection("third", 1)
	tester.Key(0, ui.KeyK)
	expectSelection("second", -1)

	tester.Key(0, ui.KeyS)
	select {
	case guid := <-starred:
		if guid != "second" {
			t.Errorf("starred guid = %q, want second", guid)
		}
	case <-time.After(time.Second):
		t.Fatal("s did not toggle the selected article's star")
	}
}

func TestPostListLoadsSelectedFeedAndNextPage(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 7, Title: "Selected feed", Unread: 2}},
		Unread: 2,
	})
	model.SetCurrentFeed(7)
	queries := make(chan app.PostQuery, 2)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			post := app.PostSummary{FeedID: 7, GUID: "first", Title: "First article"}
			if query.Offset == 1 {
				post = app.PostSummary{FeedID: 7, GUID: "second", Title: "Second article"}
			}
			return app.PostPage{Posts: []app.PostSummary{post}, Total: 2, Unread: 2}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	var first app.PostQuery
	select {
	case first = <-queries:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	if first.FeedID != 7 || first.Offset != 0 {
		t.Errorf("first query = %+v, want feed 7 offset 0", first)
	}
	tester.Frame()

	if err := tester.Click("加载更多（还有 1 篇）"); err != nil {
		t.Fatalf("click load more: %v", err)
	}
	var second app.PostQuery
	select {
	case second = <-queries:
	case <-time.After(time.Second):
		t.Fatal("loading the next page did not query posts")
	}
	if second.FeedID != 7 || second.Offset != 1 {
		t.Errorf("second query = %+v, want feed 7 offset 1", second)
	}
	tester.Frame()
	if !tester.HasText("Second article") {
		t.Error("article list is missing the loaded second page")
	}
}

func TestPostListLoadMoreCountsOnlyUnreadPosts(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 7, Title: "Selected feed", Unread: 25}},
		Unread: 25,
	})
	model.SetCurrentFeed(7)
	queries := make(chan app.PostQuery, 2)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			page := app.PostPage{Total: 25, Unread: 25}
			if query.Offset == 17 {
				page.Total = 22
				for i := 21; i <= 25; i++ {
					page.Posts = append(page.Posts, app.PostSummary{
						FeedID: 7, GUID: fmt.Sprintf("guid-%d", i), Title: fmt.Sprintf("Post %d", i),
					})
				}
				return page, nil
			}
			for i := 1; i <= 20; i++ {
				page.Posts = append(page.Posts, app.PostSummary{
					FeedID: 7, GUID: fmt.Sprintf("guid-%d", i), Title: fmt.Sprintf("Post %d", i),
				})
			}
			return page, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 1600)
	tester.Frame()
	select {
	case <-queries:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	tester.Frame()

	// Reading three listed posts only grays them in place; the rows stay.
	// Without a total adjustment the stale counter would read eight.
	for i := 1; i <= 3; i++ {
		model.SetPostUpdated(app.PostSummary{
			FeedID: 7, GUID: fmt.Sprintf("guid-%d", i), Title: fmt.Sprintf("Post %d", i), Read: true,
		})
		tester.Frame()
	}

	if err := tester.Click("加载更多（还有 5 篇）"); err != nil {
		t.Fatalf("click load more: %v", err)
	}
	var second app.PostQuery
	select {
	case second = <-queries:
	case <-time.After(time.Second):
		t.Fatal("loading the next page did not query posts")
	}
	if second.Offset != 17 {
		t.Errorf("second query offset = %d, want 17 (only unread rows count)", second.Offset)
	}
	tester.Frame()
	if !tester.HasText("Post 25") {
		t.Error("article list is missing posts after the in-place read rows")
	}
}

func TestPostListAutoLoadsNextPageNearBottom(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 7, Title: "Selected feed", Unread: 25}},
		Unread: 25,
	})
	model.SetCurrentFeed(7)
	queries := make(chan app.PostQuery, 4)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			page := app.PostPage{Total: 25, Unread: 25}
			start := query.Offset + 1
			for i := start; i <= min(start+pageSize-1, 25); i++ {
				page.Posts = append(page.Posts, app.PostSummary{
					FeedID: 7, GUID: fmt.Sprintf("guid-%d", i), Title: fmt.Sprintf("Post %d", i),
				})
			}
			return page, nil
		},
	})

	// A short viewport guarantees the content overflows on every platform;
	// otherwise short rows fit whole pages in and scrolling cannot reach the
	// bottom to trigger the auto load.
	tester := ui.NewTester(postList.View, 360, 200)
	tester.Frame()
	select {
	case first := <-queries:
		if first.Offset != 0 {
			t.Errorf("first query offset = %d, want 0", first.Offset)
		}
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	tester.Frame()

	select {
	case extra := <-queries:
		t.Fatalf("unexpected query before scrolling: %+v", extra)
	default:
	}

	postList.scroll.Y = postList.scroll.MaxY
	tester.Frame()
	select {
	case second := <-queries:
		if second.Offset != 20 {
			t.Errorf("second query offset = %d, want 20", second.Offset)
		}
	case <-time.After(time.Second):
		t.Fatal("scrolling to the bottom did not load the next page")
	}
	tester.Frame()
	if !tester.HasText("Post 25") {
		t.Error("article list is missing posts from the auto-loaded page")
	}

	// Everything is loaded now: staying at the bottom must not query again.
	postList.scroll.Y = postList.scroll.MaxY
	tester.Frame()
	select {
	case extra := <-queries:
		t.Fatalf("unexpected query after all posts loaded: %+v", extra)
	default:
	}
}

func TestPostListFilterIconButtons(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	queries := make(chan app.PostQuery, 3)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			return app.PostPage{}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	var initial app.PostQuery
	select {
	case initial = <-queries:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	if !initial.UnreadOnly || initial.StarredOnly {
		t.Fatalf("initial query = %+v, want unread only", initial)
	}

	if err := tester.Click("只看未读"); err != nil {
		t.Fatalf("click unread filter: %v", err)
	}
	var unread app.PostQuery
	select {
	case unread = <-queries:
	case <-time.After(time.Second):
		t.Fatal("clicking the unread filter did not query posts")
	}
	if unread.UnreadOnly || unread.StarredOnly {
		t.Errorf("unread query = %+v, want filters off", unread)
	}

	if err := tester.Click("只看星标"); err != nil {
		t.Fatalf("click starred filter: %v", err)
	}
	var starred app.PostQuery
	select {
	case starred = <-queries:
	case <-time.After(time.Second):
		t.Fatal("clicking the starred filter did not query posts")
	}
	if starred.UnreadOnly || !starred.StarredOnly {
		t.Errorf("starred query = %+v, want starred only", starred)
	}
}

func TestPostListMatchesWebviewTextSize(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	postList := New(model, Callbacks{})

	var fontSize float32
	tester := ui.NewTester(func(c *ui.Context) {
		fontSize = postList.Theme(c).FontSize
	}, 360, 520)
	tester.Frame()

	if fontSize != textSize {
		t.Errorf("article list font size = %v, want %v", fontSize, textSize)
	}
}

func TestPostListSearchScopeToggle(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	queries := make(chan app.PostQuery, 4)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			return app.PostPage{}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	select {
	case initial := <-queries:
		if initial.SearchScope != "" {
			t.Fatalf("initial scope = %q; want the title default", initial.SearchScope)
		}
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}

	if err := tester.Click("搜索文章"); err != nil {
		t.Fatalf("open search: %v", err)
	}

	if err := tester.Click("全文"); err != nil {
		t.Fatalf("choose fulltext scope: %v", err)
	}
	var fulltext app.PostQuery
	select {
	case fulltext = <-queries:
	case <-time.After(time.Second):
		t.Fatal("switching scope did not query posts")
	}
	if fulltext.SearchScope != app.SearchScopeFulltext {
		t.Errorf("scope = %q; want %q", fulltext.SearchScope, app.SearchScopeFulltext)
	}

	if err := tester.Click("标题"); err != nil {
		t.Fatalf("choose title scope: %v", err)
	}
	var title app.PostQuery
	select {
	case title = <-queries:
	case <-time.After(time.Second):
		t.Fatal("switching back did not query posts")
	}
	if title.SearchScope != app.SearchScopeTitle {
		t.Errorf("scope = %q; want %q", title.SearchScope, app.SearchScopeTitle)
	}
}

func TestPostListScalesLayoutWithFontSize(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeLarge, "")
	postList := New(model, Callbacks{
		List: func(app.PostQuery) (app.PostPage, error) {
			return app.PostPage{}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	button, ok := tester.Access("搜索文章")
	if !ok {
		t.Fatal("search button is missing")
	}
	want := float32(28 * 1.3)
	if button.Bounds.W < want-0.1 || button.Bounds.W > want+0.1 {
		t.Errorf("search button width = %v, want %v at large text", button.Bounds.W, want)
	}
}

func TestSelectionDirection(t *testing.T) {
	posts := []app.PostSummary{
		{FeedID: 1, GUID: "first"},
		{FeedID: 1, GUID: "second"},
		{FeedID: 1, GUID: "third"},
	}
	tests := []struct {
		name    string
		current app.PostSelection
		post    app.PostSummary
		want    int
	}{
		{"no current selection", app.PostSelection{}, posts[0], 0},
		{"later post", app.PostSelection{FeedID: 1, GUID: "first"}, posts[2], 1},
		{"earlier post", app.PostSelection{FeedID: 1, GUID: "third"}, posts[0], -1},
		{"same post", app.PostSelection{FeedID: 1, GUID: "second"}, posts[1], 0},
		{"current missing from list", app.PostSelection{FeedID: 1, GUID: "gone"}, posts[0], 0},
		{"post missing from list", app.PostSelection{FeedID: 1, GUID: "first"}, app.PostSummary{FeedID: 1, GUID: "gone"}, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := selectionDirection(posts, test.current, test.post); got != test.want {
				t.Errorf("selection direction = %d, want %d", got, test.want)
			}
		})
	}
}

func TestPostListReloadKeepsTopVisiblePost(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 7, Title: "Selected feed", Unread: 20}},
		Unread: 20,
	})
	model.SetCurrentFeed(7)

	dataset := &testDataset{}
	dataset.set(postRange(1, 20), 20)
	queries := make(chan app.PostQuery, 4)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			posts, total := dataset.page(query.Offset, query.Limit)
			return app.PostPage{Posts: posts, Total: total, Unread: total}, nil
		},
	})

	// A short viewport keeps the list overflowing on every platform, so
	// paging and scroll restoration do not depend on font metrics.
	tester := ui.NewTester(postList.View, 360, 200)
	tester.Frame()
	select {
	case <-queries:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	waitForCondition(t, tester, func() bool { return len(postList.posts) == 20 }, "initial page")
	tester.Frame()

	postList.scroll.Y = 0
	tester.Frame()
	anchor := postRowKey(app.PostSummary{FeedID: 7, GUID: "guid-1"})
	if postList.anchorKey != anchor {
		t.Fatalf("anchorKey = %q, want %q", postList.anchorKey, anchor)
	}

	// Ten fresh articles land on top; the first page keeps guid-1 at index 10.
	dataset.set(append(postRange(101, 110), postRange(1, 10)...), 30)
	model.SetPostsChanged()
	reload := frameUntilQuery(t, tester, queries, "a posts.changed reload")
	if reload.Offset != 0 || reload.Limit != pageSize {
		t.Errorf("reload query = %+v, want offset 0 limit %d", reload, pageSize)
	}
	waitForCondition(t, tester, func() bool {
		return len(postList.posts) == 20 && postList.posts[10].GUID == "guid-1"
	}, "reloaded first page")
	// The reload arms the restore with the anchor; whether the viewport then
	// scrolls depends on row metrics, which TestPostListReloadRestoresDeepScroll
	// covers end to end.
	if postList.pendingRestoreKey != anchor {
		t.Errorf("pendingRestoreKey = %q, want anchor %q", postList.pendingRestoreKey, anchor)
	}
}

func TestPostListReloadRestoresDeepScroll(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{
		Feeds:  []app.Feed{{ID: 7, Title: "Selected feed", Unread: 60}},
		Unread: 60,
	})
	model.SetCurrentFeed(7)

	dataset := &testDataset{}
	dataset.set(postRange(1, 60), 60)
	queries := make(chan app.PostQuery, 4)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			queries <- query
			posts, total := dataset.page(query.Offset, query.Limit)
			return app.PostPage{Posts: posts, Total: total, Unread: total}, nil
		},
	})

	// A short viewport keeps the list overflowing on every platform, so
	// paging and scroll restoration do not depend on font metrics.
	tester := ui.NewTester(postList.View, 360, 200)
	tester.Frame()
	select {
	case <-queries:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	waitForCondition(t, tester, func() bool { return len(postList.posts) == 20 }, "initial page")
	tester.Frame()

	// Page in the rest by scrolling to the bottom; autoLoadMore chains across
	// frames, so keep framing until every post is loaded.
	postList.scroll.Y = postList.scroll.MaxY
	for len(postList.posts) < 60 {
		loaded := len(postList.posts)
		more := frameUntilQuery(t, tester, queries, "the next page")
		if more.Offset != loaded {
			t.Errorf("next page query offset = %d, want %d", more.Offset, loaded)
		}
		waitForCondition(t, tester, func() bool { return len(postList.posts) > loaded }, "next page apply")
	}
	tester.Frame()

	postList.scroll.Y = postList.scroll.MaxY
	tester.Frame()
	savedY := postList.scroll.Y
	if savedY <= 0 {
		t.Fatalf("scroll.Y = %v, want a deep scroll position", savedY)
	}

	// The refresh refetches every loaded row in one query, so the deep anchor
	// survives and the scroll offset stays exactly where the user left it.
	// Chained autoLoadMore may still have a stale query in flight; the reload
	// wait skips those and the reload replaces the rows, dropping duplicates.
	model.SetPostsChanged()
	reload := frameUntilReloadQuery(t, tester, queries)
	if reload.Offset != 0 || reload.Limit <= pageSize {
		t.Errorf("reload query = %+v, want offset 0 and a limit covering the deep scroll", reload)
	}
	waitForCondition(t, tester, func() bool { return len(postList.posts) == reload.Limit }, "reloaded posts")
	// ScrollNearest may nudge the viewport by less than a row to fully reveal
	// the anchor; what matters is the position stayed at the deep bottom.
	waitForCondition(t, tester, func() bool {
		return postList.scroll.Y >= postList.scroll.MaxY-72
	}, "restored scroll offset")
}

// testDataset feeds the stubbed List callback; the test swaps it to simulate
// a background refresh.
type testDataset struct {
	mu    sync.Mutex
	posts []app.PostSummary
	total int
}

func (d *testDataset) set(posts []app.PostSummary, total int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.posts = posts
	d.total = total
}

func (d *testDataset) page(offset, limit int) ([]app.PostSummary, int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if offset >= len(d.posts) {
		return nil, d.total
	}
	end := min(offset+limit, len(d.posts))
	return d.posts[offset:end], d.total
}

func postRange(first, last int) []app.PostSummary {
	posts := make([]app.PostSummary, 0, last-first+1)
	for i := first; i <= last; i++ {
		posts = append(posts, app.PostSummary{
			FeedID: 7, GUID: fmt.Sprintf("guid-%d", i), Title: fmt.Sprintf("Post %d", i),
		})
	}
	return posts
}

func waitForCondition(t *testing.T, tester *ui.Tester, condition func() bool, message string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		tester.Frame()
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", message)
}

// frameUntilReloadQuery renders frames until the stub reports a reload query
// (offset 0), discarding stale next-page queries a chaining autoLoadMore may
// still have in flight.
func frameUntilReloadQuery(t *testing.T, tester *ui.Tester, queries chan app.PostQuery) app.PostQuery {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		query := frameUntilQuery(t, tester, queries, "a posts.changed reload")
		if query.Offset == 0 {
			return query
		}
	}
	t.Fatal("timed out waiting for a posts.changed reload")
	return app.PostQuery{}
}

// frameUntilQuery renders frames until the stub reports another List query,
// which is how tests drive debounced reloads without sleeping blindly.
func frameUntilQuery(t *testing.T, tester *ui.Tester, queries chan app.PostQuery, message string) app.PostQuery {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		tester.Frame()
		select {
		case query := <-queries:
			return query
		case <-time.After(25 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %s", message)
		}
	}
}

func TestPostListLoadsArticleThumbnails(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall, "")
	model.SetFeeds(app.StatusFeedsSnapshot{Unread: 3})
	loaded := make(chan struct{}, 1)
	var picture bytes.Buffer
	if err := png.Encode(&picture, image.NewRGBA(image.Rect(0, 0, 8, 8))); err != nil {
		t.Fatal(err)
	}
	requested := make(chan string, 4)
	postList := New(model, Callbacks{
		List: func(query app.PostQuery) (app.PostPage, error) {
			defer func() { loaded <- struct{}{} }()
			return app.PostPage{
				Posts: []app.PostSummary{
					{FeedID: 1, GUID: "with", Title: "With image", Image: "https://example.com/a.png"},
					{FeedID: 1, GUID: "broken", Title: "Broken image", Image: "https://example.com/broken.png"},
					{FeedID: 1, GUID: "plain", Title: "No image"},
				},
				Total:  3,
				Unread: 3,
			}, nil
		},
		Image: func(_ context.Context, source string) ([]byte, error) {
			requested <- source
			if strings.HasSuffix(source, "broken.png") {
				return nil, errors.New("gone")
			}
			return picture.Bytes(), nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	select {
	case <-loaded:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	tester.Frame()

	for range 2 {
		select {
		case <-requested:
		case <-time.After(time.Second):
			t.Fatal("article thumbnails were not requested")
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(postList.thumbs) < 2 {
		tester.Frame()
		if time.Now().After(deadline) {
			t.Fatalf("thumbnails not decoded, cache = %+v", postList.thumbs)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if postList.thumbs["https://example.com/a.png"].missing {
		t.Error("loadable image marked missing")
	}
	if !postList.thumbs["https://example.com/broken.png"].missing {
		t.Error("failed image not marked missing")
	}
	// Render the loaded state: the decoded thumbnail joins the card and the
	// missing one leaves the card text-only.
	tester.Frame()
	for _, text := range []string{"With image", "Broken image", "No image"} {
		if !tester.HasText(text) {
			t.Errorf("article list is missing %q", text)
		}
	}
	select {
	case source := <-requested:
		t.Errorf("thumbnail %q was requested again after settling", source)
	case <-time.After(50 * time.Millisecond):
	}
}
