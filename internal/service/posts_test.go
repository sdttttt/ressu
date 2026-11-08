package service

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/media"
	"github.com/sdttttt/ressu/internal/store"
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

	prefs := st.Prefs()
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
