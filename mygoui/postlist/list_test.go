package postlist

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestPostListShowsArticles(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
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
	if direction := <-directions; direction != 0 {
		t.Errorf("click selection direction = %d, want 0", direction)
	}
}

func TestPostListNavigationShortcuts(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
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

func TestPostListStarDoesNotOpenArticle(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
	loaded := make(chan struct{}, 1)
	selected := make(chan struct{}, 1)
	starred := make(chan string, 1)
	postList := New(model, Callbacks{
		List: func(app.PostQuery) (app.PostPage, error) {
			defer func() { loaded <- struct{}{} }()
			return app.PostPage{
				Posts: []app.PostSummary{
					{FeedID: 1, GUID: "first", Title: "First article", Starred: true},
				},
				Total:  1,
				Unread: 1,
			}, nil
		},
		Select: func(app.PostSummary, int) {
			selected <- struct{}{}
		},
		ToggleStar: func(_ int64, guid string) (app.PostSummary, error) {
			starred <- guid
			return app.PostSummary{GUID: guid, Starred: false}, nil
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
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
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

func TestPostListFilterIconButtons(t *testing.T) {
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
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
	model.SetAppearance(app.ThemeSystem, app.FontSizeSmall)
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
	model := app.NewStatusModel("")
	model.SetAppearance(app.ThemeSystem, app.FontSizeLarge)
	postList := New(model, Callbacks{
		List: func(app.PostQuery) (app.PostPage, error) {
			return app.PostPage{}, nil
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
