package service

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	appevent "github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/core/media"
	appstatus "github.com/sdttttt/ressu/core/status"
	"github.com/sdttttt/ressu/core/store"
)

func TestPostsListAggregatesFeeds(t *testing.T) {
	st := openStore(t)
	first, err := st.AddFeed(store.Feed{URL: "https://example.com/first", Title: "First"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := st.AddFeed(store.Feed{URL: "https://example.com/second", Title: "Second"})
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)
	for _, feed := range []store.Feed{first, second} {
		incoming := make([]store.Post, 0, 2)
		for i := range 2 {
			incoming = append(incoming, store.Post{
				GUID:        fmt.Sprintf("%d-%d", feed.ID, i),
				Title:       fmt.Sprintf("Post %d-%d", feed.ID, i),
				PublishedAt: base.Add(time.Duration(feed.ID)*time.Hour + time.Duration(i)*time.Minute),
			})
		}
		if _, err := st.MergePosts(feed.ID, incoming); err != nil {
			t.Fatal(err)
		}
	}

	page, err := NewPosts(st).List(PostQuery{FeedID: AggregateFeedID, Limit: 4})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 4 || len(page.Posts) != 4 {
		t.Fatalf("page = %+v", page)
	}
	if page.Unread != 4 {
		t.Errorf("unread = %d", page.Unread)
	}
	if page.Posts[0].GUID != "2-1" || page.Posts[2].GUID != "1-1" {
		t.Errorf("aggregate order = %+v", page.Posts)
	}

	ascending, err := NewPosts(st).List(PostQuery{
		FeedID:    AggregateFeedID,
		Limit:     4,
		SortOrder: SortOrderAscending,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ascending.Posts[0].GUID != "1-0" || ascending.Posts[3].GUID != "2-1" {
		t.Errorf("ascending order = %+v", ascending.Posts)
	}

	changed, err := NewPosts(st).MarkAllRead(AggregateFeedID)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 4 {
		t.Errorf("aggregate changed = %d", changed)
	}
}

func TestGetRewritesRemoteImagesForLocalProtocol(t *testing.T) {
	st := openStore(t)
	_, err := st.AddFeed(store.Feed{
		URL:     "https://example.com/feed.xml",
		Title:   "Example",
		SiteURL: "https://example.com/",
	})
	if err != nil {
		t.Fatal(err)
	}
	content := `<p>正文</p><img src="https://cdn.example.com/photo.png">` +
		`<img src="data:image/png;base64,aGVsbG8=">`
	if _, err := st.MergePosts(1, []store.Post{{
		FeedID:  1,
		GUID:    "photo",
		URL:     "https://example.com/photo",
		Content: content,
	}}); err != nil {
		t.Fatal(err)
	}

	articleMedia, err := media.Open(filepath.Join(t.TempDir(), "articles"), media.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer articleMedia.Close()
	mediaURL, ok := articleMedia.LocalURL("https://cdn.example.com/photo.png")
	if !ok {
		t.Fatal("source URL could not be mapped")
	}
	prefs := st.Prefs()
	prefs.RemoteResources = store.ResourceAccessGoProxy
	if err := st.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}
	post, err := NewPostsWithMedia(st, articleMedia).Get(1, "photo")
	if err != nil {
		t.Fatal(err)
	}

	if strings.Contains(post.Content, "https://cdn.example.com/photo.png") ||
		!strings.Contains(post.Content, `src="`+mediaURL+`"`) ||
		!strings.Contains(post.Content, `src="data:image/png;base64,aGVsbG8="`) {
		t.Fatalf("content = %s", post.Content)
	}
	withoutMedia, err := NewPosts(st).Get(1, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(withoutMedia.Content, "https://cdn.example.com/photo.png") {
		t.Fatalf("legacy posts service unexpectedly changed content: %s", withoutMedia.Content)
	}

	prefs = st.Prefs()
	prefs.RemoteResources = store.ResourceAccessWebview
	if err := st.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}
	webviewPost, err := NewPostsWithMedia(st, articleMedia).Get(1, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(webviewPost.Content, "https://cdn.example.com/photo.png") ||
		strings.Contains(webviewPost.Content, mediaURL) {
		t.Fatalf("Webview content = %s", webviewPost.Content)
	}
}

func TestGetUsesFeedRemoteResources(t *testing.T) {
	st := openStore(t)
	content := `<p>正文</p><img src="https://cdn.example.com/photo.png">`
	newFeedWithPost := func(id int64, url, mode string) {
		t.Helper()
		if _, err := st.AddFeed(store.Feed{ID: id, URL: url, Title: "源", RemoteResources: mode}); err != nil {
			t.Fatal(err)
		}
		if _, err := st.MergePosts(id, []store.Post{{
			FeedID:  id,
			GUID:    "photo",
			URL:     url + "/photo",
			Content: content,
		}}); err != nil {
			t.Fatal(err)
		}
	}

	// The global preference is the webview mode.
	prefs := st.Prefs()
	prefs.RemoteResources = store.ResourceAccessWebview
	if err := st.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}
	newFeedWithPost(1, "https://a.example.com/feed", store.ResourceAccessGoProxy)
	newFeedWithPost(2, "https://b.example.com/feed", store.ResourceAccessWebview)
	newFeedWithPost(3, "https://c.example.com/feed", "")

	articleMedia, err := media.Open(filepath.Join(t.TempDir(), "articles"), media.Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer articleMedia.Close()
	mediaURL, ok := articleMedia.LocalURL("https://cdn.example.com/photo.png")
	if !ok {
		t.Fatal("source URL could not be mapped")
	}
	posts := NewPostsWithMedia(st, articleMedia)

	proxied, err := posts.Get(1, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(proxied.Content, `src="`+mediaURL+`"`) {
		t.Fatalf("goProxy feed content = %s", proxied.Content)
	}

	direct, err := posts.Get(2, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(direct.Content, mediaURL) {
		t.Fatalf("webview feed content = %s", direct.Content)
	}

	inherited, err := posts.Get(3, "photo")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(inherited.Content, mediaURL) {
		t.Fatalf("feed without a mode should follow the webview preference: %s", inherited.Content)
	}
}

func addPostForReadTest(t *testing.T, st *store.Store) store.Feed {
	t.Helper()
	feed, err := st.AddFeed(store.Feed{URL: "https://example.com/feed", Title: "源"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.MergePosts(feed.ID, []store.Post{{
		GUID:        "a",
		Title:       "第一篇",
		PublishedAt: time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC),
	}}); err != nil {
		t.Fatal(err)
	}
	return feed
}

func TestGetMarksUnreadReadWhenPreferenceOn(t *testing.T) {
	st := openStore(t)
	feed := addPostForReadTest(t, st)

	var updated []store.PostSummary
	hub := appevent.NewHub(appstatus.New("test"), func(e appevent.Event) error {
		if e.EID == appevent.EventPostUpdated {
			updated = append(updated, *e.Post)
		}
		return nil
	})
	posts := NewPostsWithHub(st, hub)

	post, err := posts.Get(feed.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !post.Read {
		t.Error("Get should mark the unread post read")
	}
	stored, err := st.Post(feed.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Read {
		t.Error("the stored post should be read")
	}
	if st.Unread() != 0 {
		t.Errorf("store unread = %d", st.Unread())
	}
	if len(updated) != 1 || !updated[0].Read {
		t.Errorf("post.updated events = %+v", updated)
	}

	// Reading an already read post publishes nothing.
	if _, err := posts.Get(feed.ID, "a"); err != nil {
		t.Fatal(err)
	}
	if len(updated) != 1 {
		t.Errorf("post.updated events after re-read = %+v", updated)
	}
}

func TestGetKeepsUnreadWhenPreferenceOff(t *testing.T) {
	st := openStore(t)
	feed := addPostForReadTest(t, st)
	prefs := st.Prefs()
	prefs.MarkReadOnOpen = false
	if err := st.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}

	post, err := NewPosts(st).Get(feed.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if post.Read {
		t.Error("Get should keep the post unread when the preference is off")
	}
	if st.Unread() != 1 {
		t.Errorf("store unread = %d", st.Unread())
	}
}

func TestToggleReadFlipsState(t *testing.T) {
	st := openStore(t)
	feed := addPostForReadTest(t, st)

	var updated []store.PostSummary
	hub := appevent.NewHub(appstatus.New("test"), func(e appevent.Event) error {
		if e.EID == appevent.EventPostUpdated {
			updated = append(updated, *e.Post)
		}
		return nil
	})
	posts := NewPostsWithHub(st, hub)

	summary, err := posts.ToggleRead(feed.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !summary.Read {
		t.Error("ToggleRead should mark the post read")
	}
	summary, err = posts.ToggleRead(feed.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if summary.Read {
		t.Error("ToggleRead should mark the post unread again")
	}
	if st.Unread() != 1 {
		t.Errorf("store unread = %d", st.Unread())
	}
	if len(updated) != 2 || updated[0].Read == updated[1].Read {
		t.Errorf("post.updated events = %+v", updated)
	}
}
