package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sdttttt/ressu/internal/store"
)

func feedXML(items ...string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel>
<title>测试源</title>
<link>https://example.com/</link>
<description>一个测试用的订阅源</description>
` + strings.Join(items, "\n") + `
</channel></rss>`
}

func itemXML(guid, title string) string {
	return fmt.Sprintf(`<item>
  <title>%s</title>
  <link>https://example.com/%s</link>
  <guid>%s</guid>
  <pubDate>Wed, 02 Oct 2024 12:00:00 GMT</pubDate>
  <description>&lt;p&gt;%s 的摘要&lt;/p&gt;</description>
</item>`, title, guid, guid, title)
}

func openStore(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

func TestAddRefreshAndRead(t *testing.T) {
	var body atomic.Value
	body.Store(feedXML(itemXML("a", "第一篇"), itemXML("b", "第二篇")))
	var version atomic.Value
	version.Store(`"v1"`)

	var notModified int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		etag := version.Load().(string)
		if r.Header.Get("If-None-Match") == etag {
			atomic.AddInt32(&notModified, 1)
			w.WriteHeader(http.StatusNotModified)
			return
		}
		w.Header().Set("ETag", etag)
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = io.WriteString(w, body.Load().(string))
	}))
	defer server.Close()

	st := openStore(t)
	feeds := NewFeeds(st)
	posts := NewPosts(st)

	ctx := context.Background()
	feed, err := feeds.Add(ctx, server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Title != "测试源" || feed.URL != server.URL {
		t.Errorf("feed = %+v", feed)
	}
	if feed.ETag != `"v1"` {
		t.Errorf("etag = %q", feed.ETag)
	}
	if feed.Unread != 2 {
		t.Errorf("unread = %d", feed.Unread)
	}

	page, err := posts.List(PostQuery{FeedID: feed.ID})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || len(page.Posts) != 2 {
		t.Fatalf("page = %+v", page)
	}
	if page.Posts[0].Summary != "第一篇 的摘要" {
		t.Errorf("summary = %q", page.Posts[0].Summary)
	}

	full, err := posts.Get(feed.ID, "a")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(full.Content, "第一篇 的摘要") {
		t.Errorf("content = %q", full.Content)
	}

	// A refresh with the same validators gets a 304 and keeps the posts.
	summary, err := feeds.RefreshAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Updated != 1 || summary.Failed != 0 || summary.New != 0 {
		t.Errorf("summary = %+v", summary)
	}
	if atomic.LoadInt32(&notModified) != 1 {
		t.Errorf("the server was not asked with If-None-Match")
	}

	// A new item shows up after the feed changes.
	body.Store(feedXML(itemXML("a", "第一篇"), itemXML("b", "第二篇"), itemXML("c", "第三篇")))
	version.Store(`"v2"`)
	summary, err = feeds.RefreshAll(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if summary.New != 1 || summary.Updated != 1 {
		t.Errorf("summary = %+v", summary)
	}

	feed, err = st.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Unread != 3 {
		t.Errorf("unread = %d", feed.Unread)
	}

	// Marking one read lowers the count.
	if err := posts.MarkRead(feed.ID, "a", true); err != nil {
		t.Fatal(err)
	}
	feed, err = st.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Unread != 2 {
		t.Errorf("unread after read = %d", feed.Unread)
	}

	// The unread filter only lists the unread ones.
	page, err = posts.List(PostQuery{FeedID: feed.ID, UnreadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 2 || page.Unread != 2 {
		t.Errorf("unread page = %+v", page)
	}

	// Marking everything read empties the unread count.
	changed, err := posts.MarkAllRead(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 2 {
		t.Errorf("changed = %d", changed)
	}
	if st.Unread() != 0 {
		t.Errorf("store unread = %d", st.Unread())
	}
}

func TestAddRejectsWhatIsNotAFeed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "<html><body>hello</body></html>")
	}))
	defer server.Close()

	st := openStore(t)
	feeds := NewFeeds(st)

	if _, err := feeds.Add(context.Background(), server.URL); err == nil {
		t.Fatal("a page that is not a feed should be refused")
	}
	if len(st.Feeds()) != 0 {
		t.Error("nothing should be stored")
	}

	if _, err := feeds.Add(context.Background(), "not a url"); err == nil {
		t.Fatal("an invalid address should be refused")
	}
}

func TestAddRejectsDuplicates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, feedXML(itemXML("a", "第一篇")))
	}))
	defer server.Close()

	st := openStore(t)
	feeds := NewFeeds(st)

	if _, err := feeds.Add(context.Background(), server.URL); err != nil {
		t.Fatal(err)
	}
	if _, err := feeds.Add(context.Background(), server.URL); err == nil {
		t.Fatal("the same URL should not be added twice")
	}
}

func TestListCarriesTotalUnread(t *testing.T) {
	st := openStore(t)
	feeds := NewFeeds(st)
	posts := NewPosts(st)

	feed, err := st.AddFeed(store.Feed{URL: "https://example.com/feed", Title: "测试源"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.MergePosts(feed.ID, []store.Post{
		{GUID: "a", Title: "第一篇", PublishedAt: time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)},
		{GUID: "b", Title: "第二篇", PublishedAt: time.Date(2024, 10, 1, 13, 0, 0, 0, time.UTC)},
	}); err != nil {
		t.Fatal(err)
	}

	snapshot := feeds.List()
	if len(snapshot.Feeds) != 1 || snapshot.Unread != 2 {
		t.Fatalf("list = %+v", snapshot)
	}

	if _, err := posts.MarkAllRead(feed.ID); err != nil {
		t.Fatal(err)
	}
	if snapshot = feeds.List(); snapshot.Unread != 0 {
		t.Errorf("unread after read = %d", snapshot.Unread)
	}
}

func TestRefreshFailsWithoutLosingTheFeed(t *testing.T) {
	var failing atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failing.Load() {
			http.Error(w, "boom", http.StatusInternalServerError)
			return
		}
		_, _ = io.WriteString(w, feedXML(itemXML("a", "第一篇")))
	}))
	defer server.Close()

	st := openStore(t)
	feeds := NewFeeds(st)

	feed, err := feeds.Add(context.Background(), server.URL)
	if err != nil {
		t.Fatal(err)
	}

	failing.Store(true)
	if _, err := feeds.Refresh(context.Background(), feed.ID); err == nil {
		t.Fatal("the refresh should fail")
	}

	stored, err := st.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.LastError == "" {
		t.Error("the error should be kept on the feed")
	}
	if len(st.Feeds()) != 1 {
		t.Error("the feed should stay subscribed")
	}

	posts, err := st.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Errorf("posts = %d", len(posts))
	}
}

func TestPostsListPagesAndSearches(t *testing.T) {
	st := openStore(t)
	feed, err := st.AddFeed(store.Feed{URL: "https://example.com/feed", Title: "源"})
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)
	incoming := make([]store.Post, 0, 5)
	for i := range 5 {
		incoming = append(incoming, store.Post{
			GUID:        fmt.Sprintf("p%d", i),
			Title:       fmt.Sprintf("文章 %d", i),
			PublishedAt: base.Add(time.Duration(i) * time.Hour),
		})
	}
	if _, err := st.MergePosts(feed.ID, incoming); err != nil {
		t.Fatal(err)
	}

	posts := NewPosts(st)

	page, err := posts.List(PostQuery{FeedID: feed.ID, Limit: 2})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 5 || len(page.Posts) != 2 {
		t.Fatalf("page = %+v", page)
	}
	if page.Posts[0].Title != "文章 4" {
		t.Errorf("first post = %q", page.Posts[0].Title)
	}

	page, err = posts.List(PostQuery{FeedID: feed.ID, Limit: 2, Offset: 4})
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Posts) != 1 || page.Posts[0].Title != "文章 0" {
		t.Errorf("second page = %+v", page)
	}

	page, err = posts.List(PostQuery{FeedID: feed.ID, Search: "文章 3"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Errorf("search total = %d", page.Total)
	}

	starred, err := posts.ToggleStar(feed.ID, "p2")
	if err != nil {
		t.Fatal(err)
	}
	if !starred.Starred {
		t.Error("the post should be starred")
	}
	page, err = posts.List(PostQuery{FeedID: feed.ID, StarredOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || page.Posts[0].GUID != "p2" {
		t.Errorf("starred page = %+v", page)
	}
}

func TestPostsListSearchesTitlesOnly(t *testing.T) {
	st := openStore(t)
	feed, err := st.AddFeed(store.Feed{URL: "https://example.com/feed", Title: "源"})
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)
	if _, err := st.MergePosts(feed.ID, []store.Post{{
		GUID:        "needle",
		Title:       "标题匹配",
		Summary:     "摘要 keyword",
		Author:      "作者 keyword",
		Content:     "<p>正文 keyword</p>",
		PublishedAt: base,
	}}); err != nil {
		t.Fatal(err)
	}

	posts := NewPosts(st)
	page, err := posts.List(PostQuery{FeedID: feed.ID, Search: "keyword"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 {
		t.Errorf("non-title search total = %d", page.Total)
	}

	page, err = posts.List(PostQuery{FeedID: feed.ID, Search: "标题"})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 {
		t.Errorf("title search total = %d", page.Total)
	}
}
