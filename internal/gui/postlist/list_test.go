package postlist

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/service"
	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

func TestPostListShowsArticles(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
	model.SetFeeds(appstatus.FeedsSnapshot{Unread: 1})
	loaded := make(chan struct{}, 1)
	selected := make(chan store.PostSummary, 1)
	now := time.Now()
	postList := New(model, Callbacks{
		List: func(query service.PostQuery) (service.PostPage, error) {
			defer func() {
				loaded <- struct{}{}
			}()
			return service.PostPage{
				Posts: []store.PostSummary{
					{
						FeedID:      1,
						GUID:        "first",
						Title:       "First article",
						Summary:     "First summary",
						Author:      "Author",
						WordCount:   1234,
						PublishedAt: now,
					},
					{FeedID: 1, GUID: "second", Title: "Second article", Summary: "Second summary", Read: true},
				},
				Total:  2,
				Unread: 1,
			}, nil
		},
		Select: func(post store.PostSummary) {
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
		"Author",
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
}

func TestPostListNavigationShortcuts(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
	model.SetFeeds(appstatus.FeedsSnapshot{Unread: 2})
	loaded := make(chan struct{}, 1)
	selected := make(chan store.PostSummary, 8)
	starred := make(chan string, 1)
	postList := New(model, Callbacks{
		List: func(service.PostQuery) (service.PostPage, error) {
			defer func() { loaded <- struct{}{} }()
			return service.PostPage{
				Posts: []store.PostSummary{
					{FeedID: 1, GUID: "first", Title: "First article"},
					{FeedID: 1, GUID: "second", Title: "Second article", Read: true},
					{FeedID: 1, GUID: "third", Title: "Third article"},
				},
				Total:  3,
				Unread: 2,
			}, nil
		},
		Select: func(post store.PostSummary) {
			selection := appstatus.PostSelection{FeedID: post.FeedID, GUID: post.GUID}
			model.SetCurrentPost(selection, post)
			selected <- post
		},
		ToggleStar: func(_ int64, guid string) (store.PostSummary, error) {
			starred <- guid
			return store.PostSummary{}, nil
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

	expectSelection := func(guid string) {
		t.Helper()
		select {
		case post := <-selected:
			if post.GUID != guid {
				t.Errorf("selected guid = %q, want %q", post.GUID, guid)
			}
		case <-time.After(time.Second):
			t.Fatalf("pressing a shortcut did not select %q", guid)
		}
	}

	tester.Key(0, ui.KeyJ)
	expectSelection("first")
	tester.Key(0, ui.KeyJ)
	expectSelection("second")
	tester.Key(0, ui.KeyN)
	expectSelection("third")
	tester.Key(0, ui.KeyK)
	expectSelection("second")

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

func TestPostListStarDoesNotOpenArticle(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
	loaded := make(chan struct{}, 1)
	selected := make(chan struct{}, 1)
	starred := make(chan string, 1)
	postList := New(model, Callbacks{
		List: func(service.PostQuery) (service.PostPage, error) {
			defer func() { loaded <- struct{}{} }()
			return service.PostPage{
				Posts: []store.PostSummary{
					{FeedID: 1, GUID: "first", Title: "First article", Starred: true},
				},
				Total:  1,
				Unread: 1,
			}, nil
		},
		Select: func(store.PostSummary) {
			selected <- struct{}{}
		},
		ToggleStar: func(_ int64, guid string) (store.PostSummary, error) {
			starred <- guid
			return store.PostSummary{GUID: guid, Starred: false}, nil
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

	row, ok := tester.Access("First article")
	if !ok {
		t.Fatal("article row is missing")
	}
	tester.ClickAt(row.Bounds.X+row.Bounds.W-16, row.Bounds.Y+row.Bounds.H-16)
	select {
	case guid := <-starred:
		if guid != "first" {
			t.Errorf("toggled star for %q, want first", guid)
		}
	case <-time.After(time.Second):
		t.Fatal("clicking the star did not toggle it")
	}
	select {
	case <-selected:
		t.Fatal("clicking the star opened the article")
	default:
	}
}

func TestPostListLoadsSelectedFeedAndNextPage(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
	model.SetFeeds(appstatus.FeedsSnapshot{
		Feeds:  []store.Feed{{ID: 7, Title: "Selected feed", Unread: 2}},
		Unread: 2,
	})
	model.SetCurrentFeed(7)
	queries := make(chan service.PostQuery, 2)
	postList := New(model, Callbacks{
		List: func(query service.PostQuery) (service.PostPage, error) {
			queries <- query
			post := store.PostSummary{FeedID: 7, GUID: "first", Title: "First article"}
			if query.Offset == 1 {
				post = store.PostSummary{FeedID: 7, GUID: "second", Title: "Second article"}
			}
			return service.PostPage{Posts: []store.PostSummary{post}, Total: 2, Unread: 2}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	var first service.PostQuery
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
	var second service.PostQuery
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

func TestPostListFilterIconButtons(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
	queries := make(chan service.PostQuery, 3)
	postList := New(model, Callbacks{
		List: func(query service.PostQuery) (service.PostPage, error) {
			queries <- query
			return service.PostPage{}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	var initial service.PostQuery
	select {
	case initial = <-queries:
	case <-time.After(time.Second):
		t.Fatal("article list did not load")
	}
	if initial.UnreadOnly || initial.StarredOnly {
		t.Fatalf("initial query = %+v, want filters off", initial)
	}

	if err := tester.Click("只看未读"); err != nil {
		t.Fatalf("click unread filter: %v", err)
	}
	var unread service.PostQuery
	select {
	case unread = <-queries:
	case <-time.After(time.Second):
		t.Fatal("clicking the unread filter did not query posts")
	}
	if !unread.UnreadOnly || unread.StarredOnly {
		t.Errorf("unread query = %+v, want unread only", unread)
	}

	if err := tester.Click("只看星标"); err != nil {
		t.Fatalf("click starred filter: %v", err)
	}
	var starred service.PostQuery
	select {
	case starred = <-queries:
	case <-time.After(time.Second):
		t.Fatal("clicking the starred filter did not query posts")
	}
	if !starred.UnreadOnly || !starred.StarredOnly {
		t.Errorf("starred query = %+v, want both filters on", starred)
	}
}

func TestPostListMatchesWebviewTextSize(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeSmall)
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

func TestPostListScalesLayoutWithFontSize(t *testing.T) {
	model := appstatus.New("")
	model.SetAppearance(store.ThemeSystem, store.FontSizeLarge)
	postList := New(model, Callbacks{
		List: func(service.PostQuery) (service.PostPage, error) {
			return service.PostPage{}, nil
		},
	})

	tester := ui.NewTester(postList.View, 360, 520)
	tester.Frame()
	button, ok := tester.Access("搜索标题")
	if !ok {
		t.Fatal("search button is missing")
	}
	want := float32(28 * 1.3)
	if button.Bounds.W < want-0.1 || button.Bounds.W > want+0.1 {
		t.Errorf("search button width = %v, want %v at large text", button.Bounds.W, want)
	}
}
