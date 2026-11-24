package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	bolt "go.etcd.io/bbolt"
)

var base = time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestOpenCreatesDatabase(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	if _, err := os.Stat(filepath.Join(dir, dbFileName)); err != nil {
		t.Errorf("%s: %v", dbFileName, err)
	}
	if got := s.Dir(); got != dir {
		t.Errorf("dir = %q, want %q", got, dir)
	}
	if got := s.Prefs().PostsPerFeed; got != DefaultPostsPerFeed {
		t.Errorf("posts per feed = %d", got)
	}
	if len(s.Feeds()) != 0 {
		t.Error("a new store should have no feeds")
	}
}

func TestDBSizeMatchesFile(t *testing.T) {
	s := openTemp(t)

	info, err := os.Stat(filepath.Join(s.Dir(), dbFileName))
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.DBSize()
	if err != nil {
		t.Fatal(err)
	}
	if got != info.Size() {
		t.Errorf("DBSize() = %d, want file size %d", got, info.Size())
	}
}

func TestPrefsNormalizeRemoteResources(t *testing.T) {
	if prefs := DefaultPrefs(); prefs.RemoteResources != ResourceAccessWebview {
		t.Errorf("default remote resources = %q, want %q", prefs.RemoteResources, ResourceAccessWebview)
	}

	prefs := Prefs{RemoteResources: "unexpected"}
	prefs.Normalize()
	if prefs.RemoteResources != ResourceAccessGoProxy {
		t.Errorf("unknown remote resources = %q, want %q", prefs.RemoteResources, ResourceAccessGoProxy)
	}

	prefs = Prefs{RemoteResources: ResourceAccessGoProxy}
	prefs.Normalize()
	if prefs.RemoteResources != ResourceAccessGoProxy {
		t.Errorf("goProxy remote resources = %q", prefs.RemoteResources)
	}

	prefs = Prefs{RemoteResources: ResourceAccessWebview}
	prefs.Normalize()
	if prefs.RemoteResources != ResourceAccessWebview {
		t.Errorf("webview remote resources = %q", prefs.RemoteResources)
	}
}

func TestPrefsNormalizePostSort(t *testing.T) {
	if prefs := DefaultPrefs(); prefs.PostSort != PostSortPublished {
		t.Errorf("default post sort = %q, want %q", prefs.PostSort, PostSortPublished)
	}

	prefs := Prefs{PostSort: PostSortFetched}
	prefs.Normalize()
	if prefs.PostSort != PostSortFetched {
		t.Errorf("fetched post sort = %q", prefs.PostSort)
	}

	prefs = Prefs{PostSort: ""}
	prefs.Normalize()
	if prefs.PostSort != PostSortPublished {
		t.Errorf("empty post sort = %q, want %q", prefs.PostSort, PostSortPublished)
	}
}

func TestAddFeedRejectsDuplicates(t *testing.T) {
	s := openTemp(t)

	first, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != 1 {
		t.Errorf("first id = %d", first.ID)
	}

	if _, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "Again"}); !errors.Is(err, ErrDuplicate) {
		t.Errorf("duplicate error = %v", err)
	}

	second, err := s.AddFeed(Feed{URL: "https://b.example/feed", Title: "B"})
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != 2 {
		t.Errorf("second id = %d", second.ID)
	}
}

func TestOpenSanitizesStoredPostHTML(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		return put(tx.Bucket([]byte(settingsBucket)), []byte(schemaVersionKey), 1)
	})
	if err != nil {
		t.Fatal(err)
	}
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.MergePosts(feed.ID, []Post{{
		GUID:        "raw",
		URL:         "https://a.example/post",
		Title:       "Raw",
		Content:     `<p>正文</p><script>alert(1)</script><a href="/next">next</a>`,
		PublishedAt: base,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	posts, err := reopened.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 {
		t.Fatalf("posts = %+v", posts)
	}
	content := posts[0].Content
	if strings.Contains(content, "<script") || strings.Contains(content, "alert(1)") {
		t.Fatalf("script survived migration: %s", content)
	}
	if !strings.Contains(content, `href="https://a.example/next" rel="noopener noreferrer" target="_blank"`) {
		t.Fatalf("links were not completed during migration: %s", content)
	}

	var version int
	err = reopened.db.View(func(tx *bolt.Tx) error {
		value := tx.Bucket([]byte(settingsBucket)).Get([]byte(schemaVersionKey))
		return json.Unmarshal(value, &version)
	})
	if err != nil {
		t.Fatal(err)
	}
	if version != currentSchemaVersion {
		t.Errorf("schema version = %d, want %d", version, currentSchemaVersion)
	}
}

func TestOpenBackfillsPostWordCounts(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.MergePosts(feed.ID, []Post{{
		GUID:        "counted",
		Title:       "Counted",
		Summary:     "摘要",
		Content:     `<p>中文 two</p>`,
		PublishedAt: base,
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = s.db.Update(func(tx *bolt.Tx) error {
		return put(tx.Bucket([]byte(settingsBucket)), []byte(schemaVersionKey), 3)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	posts, err := reopened.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].WordCount != 3 {
		t.Fatalf("posts = %+v, word count = 3", posts)
	}
	summary := posts[0].ToSummary()
	if summary.WordCount != 3 {
		t.Errorf("summary word count = %d, want 3", summary.WordCount)
	}
}

func TestOpenRecomputesUnreadCounts(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(feed.ID, []Post{
		{GUID: "a", Title: "A", PublishedAt: base},
		{GUID: "b", Title: "B", PublishedAt: base.Add(time.Hour)},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdatePost(feed.ID, "a", func(post *Post) { post.Read = true }); err != nil {
		t.Fatal(err)
	}
	if err := s.db.Update(func(tx *bolt.Tx) error {
		// Simulate the stale count a pre-fix build could persist; opening
		// the store must derive the count from the posts instead of
		// trusting the cached value.
		feeds := tx.Bucket([]byte(feedsBucket))
		var f Feed
		if _, err := get(feeds, feedKey(feed.ID), &f); err != nil {
			return err
		}
		f.Unread = 2
		return put(feeds, feedKey(feed.ID), f)
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	updated, err := reopened.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Unread != 1 {
		t.Errorf("unread = %d, want 1", updated.Unread)
	}
	if got := reopened.Unread(); got != 1 {
		t.Errorf("total unread = %d, want 1", got)
	}
}

func TestUpdateFeedPreservesUnread(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(feed.ID, []Post{
		{GUID: "a", Title: "A", PublishedAt: base},
		{GUID: "b", Title: "B", PublishedAt: base.Add(time.Hour)},
	}); err != nil {
		t.Fatal(err)
	}

	// A caller that read the feed before posts changed holds a stale count;
	// UpdateFeed must not write it back.
	stale, err := s.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdatePost(feed.ID, "a", func(post *Post) { post.Read = true }); err != nil {
		t.Fatal(err)
	}
	if err := s.UpdateFeed(stale); err != nil {
		t.Fatal(err)
	}

	updated, err := s.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Unread != 1 {
		t.Errorf("unread = %d, want 1", updated.Unread)
	}
	if got := s.Unread(); got != 1 {
		t.Errorf("total unread = %d, want 1", got)
	}
}

func TestMergePostsKeepsReadAndStarred(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}

	added, err := s.MergePosts(feed.ID, []Post{
		{GUID: "a", Title: "A", PublishedAt: base},
		{GUID: "b", Title: "B", PublishedAt: base.Add(time.Hour)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if added != 2 {
		t.Errorf("added = %d", added)
	}
	if got := s.Unread(); got != 2 {
		t.Errorf("unread = %d", got)
	}

	if _, err := s.UpdatePost(feed.ID, "a", func(post *Post) {
		post.Read = true
		post.Starred = true
	}); err != nil {
		t.Fatal(err)
	}

	added, err = s.MergePosts(feed.ID, []Post{
		{GUID: "a", Title: "A changed", PublishedAt: base},
		{GUID: "c", Title: "C", PublishedAt: base.Add(2 * time.Hour)},
	})
	if err != nil {
		t.Fatal(err)
	}
	if added != 1 {
		t.Errorf("added = %d", added)
	}

	posts, err := s.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 3 {
		t.Fatalf("posts = %d", len(posts))
	}

	want := []string{"c", "b", "a"}
	for i, guid := range want {
		if posts[i].GUID != guid {
			t.Errorf("post %d = %q, want %q", i, posts[i].GUID, guid)
		}
	}

	updated := posts[2]
	if updated.Title != "A changed" {
		t.Errorf("title = %q", updated.Title)
	}
	if !updated.Read || !updated.Starred {
		t.Errorf("state was lost: read=%v starred=%v", updated.Read, updated.Starred)
	}
	if got := s.Unread(); got != 2 {
		t.Errorf("unread = %d", got)
	}

	feed, err = s.Feed(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Unread != 2 {
		t.Errorf("feed unread = %d", feed.Unread)
	}
}

func TestMergePostsTrimsToLimit(t *testing.T) {
	s := openTemp(t)
	prefs := s.Prefs()
	prefs.PostsPerFeed = 50
	if err := s.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}

	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}

	posts := make([]Post, 0, 60)
	for i := range 60 {
		posts = append(posts, Post{
			GUID:        fmt.Sprintf("post-%d", i),
			PublishedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}
	if _, err := s.MergePosts(feed.ID, posts); err != nil {
		t.Fatal(err)
	}

	kept, err := s.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 50 {
		t.Fatalf("kept = %d", len(kept))
	}
	if kept[0].GUID != "post-59" {
		t.Errorf("newest kept = %q", kept[0].GUID)
	}
	if last := kept[len(kept)-1].GUID; last != "post-10" {
		t.Errorf("oldest kept = %q", last)
	}
}

func TestMergePostsTrimsToFeedKeepCount(t *testing.T) {
	s := openTemp(t)
	prefs := s.Prefs()
	prefs.PostsPerFeed = 50
	if err := s.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}

	limited, err := s.AddFeed(Feed{URL: "https://a.example/feed", KeepCount: 60})
	if err != nil {
		t.Fatal(err)
	}
	following, err := s.AddFeed(Feed{URL: "https://b.example/feed"})
	if err != nil {
		t.Fatal(err)
	}

	posts := make([]Post, 0, 80)
	for i := range 80 {
		posts = append(posts, Post{
			GUID:        fmt.Sprintf("post-%d", i),
			PublishedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}
	for _, feed := range []Feed{limited, following} {
		if _, err := s.MergePosts(feed.ID, posts); err != nil {
			t.Fatal(err)
		}
	}

	kept, err := s.Posts(limited.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 60 {
		t.Errorf("feed keep count kept = %d", len(kept))
	}
	if kept[0].GUID != "post-79" {
		t.Errorf("newest kept = %q", kept[0].GUID)
	}

	kept, err = s.Posts(following.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 50 {
		t.Errorf("global limit kept = %d", len(kept))
	}
}

func TestMergePostsTrimsKeepsStarred(t *testing.T) {
	s := openTemp(t)
	prefs := s.Prefs()
	prefs.PostsPerFeed = 50
	if err := s.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}

	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}

	posts := make([]Post, 0, 60)
	for i := range 60 {
		posts = append(posts, Post{
			GUID:        fmt.Sprintf("post-%d", i),
			PublishedAt: base.Add(time.Duration(i) * time.Minute),
		})
	}
	posts[0].Starred = true // oldest post, beyond the limit
	if _, err := s.MergePosts(feed.ID, posts); err != nil {
		t.Fatal(err)
	}

	kept, err := s.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(kept) != 51 {
		t.Fatalf("kept = %d", len(kept))
	}
	if kept[len(kept)-1].GUID != "post-0" || !kept[len(kept)-1].Starred {
		t.Errorf("oldest kept = %q, starred = %v", kept[len(kept)-1].GUID, kept[len(kept)-1].Starred)
	}
	if kept[0].GUID != "post-59" {
		t.Errorf("newest kept = %q", kept[0].GUID)
	}

	// The starred post survives a later merge that pushes it even further
	// beyond the limit.
	if _, err := s.MergePosts(feed.ID, []Post{{
		GUID:        "post-60",
		PublishedAt: base.Add(60 * time.Minute),
	}}); err != nil {
		t.Fatal(err)
	}
	kept, err = s.Posts(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if kept[len(kept)-1].GUID != "post-0" {
		t.Errorf("starred dropped after later merge, oldest kept = %q", kept[len(kept)-1].GUID)
	}
}

func TestMarkAllRead(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(feed.ID, []Post{
		{GUID: "a", PublishedAt: base},
		{GUID: "b", PublishedAt: base},
	}); err != nil {
		t.Fatal(err)
	}

	changed, err := s.MarkAllRead(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 2 {
		t.Errorf("changed = %d", changed)
	}
	if got := s.Unread(); got != 0 {
		t.Errorf("unread = %d", got)
	}

	changed, err = s.MarkAllRead(feed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if changed != 0 {
		t.Errorf("second call changed = %d", changed)
	}
}

func TestSearchPostGUIDs(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}
	other, err := s.AddFeed(Feed{URL: "https://b.example/feed"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(feed.ID, []Post{
		{GUID: "title-hit", Title: "Rust 发布说明", Content: "<p>无关正文</p>", PublishedAt: base},
		{GUID: "body-hit", Title: "周记", Content: "<p>这周学了 Rust  borrow check</p>", PublishedAt: base},
		{GUID: "miss", Title: "另一篇", Content: "<p>什么都没有</p>", PublishedAt: base},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(other.ID, []Post{
		{GUID: "cross", Title: "Rust 别处", Content: "<p>别的源</p>", PublishedAt: base},
	}); err != nil {
		t.Fatal(err)
	}

	guids, err := s.SearchPostGUIDs(feed.ID, "rust")
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(guids)
	if want := []string{"body-hit", "title-hit"}; !slices.Equal(guids, want) {
		t.Errorf("guids = %v; want %v", guids, want)
	}

	// Case-insensitivity reaches case-folded text.
	guids, err = s.SearchPostGUIDs(feed.ID, "RUST")
	if err != nil {
		t.Fatal(err)
	}
	if len(guids) != 2 {
		t.Errorf("uppercase needle guids = %v; want 2 hits", guids)
	}

	// A caseless-script needle hits the body directly.
	guids, err = s.SearchPostGUIDs(feed.ID, "borrow")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(guids, []string{"body-hit"}) {
		t.Errorf("body guids = %v; want [body-hit]", guids)
	}

	guids, err = s.SearchPostGUIDs(feed.ID, "不存在")
	if err != nil {
		t.Fatal(err)
	}
	if len(guids) != 0 {
		t.Errorf("miss guids = %v; want none", guids)
	}

	guids, err = s.SearchPostGUIDs(9999, "rust")
	if err != nil {
		t.Fatal(err)
	}
	if len(guids) != 0 {
		t.Errorf("unknown feed guids = %v; want none", guids)
	}
}

func TestRemoveFeedDeletesPosts(t *testing.T) {
	s := openTemp(t)
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(feed.ID, []Post{{GUID: "a", PublishedAt: base}}); err != nil {
		t.Fatal(err)
	}

	if err := s.RemoveFeed(feed.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Feed(feed.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("feed error = %v", err)
	}
	if _, err := s.Post(feed.ID, "a"); !errors.Is(err, ErrNotFound) {
		t.Errorf("post error = %v", err)
	}
	if len(s.Feeds()) != 0 {
		t.Error("feeds should be empty")
	}
}

func TestEverythingSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	feed, err := s.AddFeed(Feed{URL: "https://a.example/feed", Title: "A"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.MergePosts(feed.ID, []Post{{GUID: "a", Title: "One", PublishedAt: base}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdatePost(feed.ID, "a", func(post *Post) {
		post.Read = true
		post.Starred = true
	}); err != nil {
		t.Fatal(err)
	}
	prefs := s.Prefs()
	prefs.Theme = ThemeDark
	prefs.PostsPerFeed = 100
	prefs.PaneSizes = []int{40, 55}
	prefs.SidebarCollapsed = true
	if err := s.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()

	got := reopened.Prefs()
	if got.Theme != ThemeDark || got.PostsPerFeed != 100 || !got.SidebarCollapsed {
		t.Errorf("prefs = %+v", got)
	}
	if len(got.PaneSizes) != 2 || got.PaneSizes[0] != 40 {
		t.Errorf("pane sizes = %v", got.PaneSizes)
	}

	feeds := reopened.Feeds()
	if len(feeds) != 1 || feeds[0].Title != "A" {
		t.Fatalf("feeds = %+v", feeds)
	}
	if feeds[0].Unread != 0 {
		t.Errorf("unread = %d", feeds[0].Unread)
	}
	posts, err := reopened.Posts(feeds[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].Title != "One" {
		t.Fatalf("posts = %+v", posts)
	}
	if !posts[0].Read || !posts[0].Starred {
		t.Errorf("flags = read:%v starred:%v", posts[0].Read, posts[0].Starred)
	}
}

func TestPrefsNormalize(t *testing.T) {
	p := Prefs{Theme: "solarized", FontSize: "huge", OpenLinks: "carrier-pigeon", PostSort: "creation", PostsPerFeed: 3, PaneSizes: []int{10}}
	p.SidebarCollapsed = true
	p.Normalize()

	if p.Theme != ThemeSystem {
		t.Errorf("theme = %q", p.Theme)
	}
	if p.FontSize != FontSizeMedium {
		t.Errorf("font size = %q", p.FontSize)
	}
	if p.OpenLinks != OpenLinksBrowser {
		t.Errorf("open links = %q", p.OpenLinks)
	}
	if p.PostSort != PostSortPublished {
		t.Errorf("post sort = %q", p.PostSort)
	}
	if p.PostsPerFeed != DefaultPostsPerFeed {
		t.Errorf("posts per feed = %d", p.PostsPerFeed)
	}
	if p.ImageCacheLimit != DefaultImageCacheLimit {
		t.Errorf("image cache limit = %d", p.ImageCacheLimit)
	}
	if len(p.PaneSizes) != 2 {
		t.Errorf("pane sizes = %v", p.PaneSizes)
	}
	if !p.SidebarCollapsed {
		t.Errorf("sidebar collapsed = false")
	}
}

func feedIDOrder(feeds []Feed) []int64 {
	ids := make([]int64, len(feeds))
	for i, f := range feeds {
		ids[i] = f.ID
	}
	return ids
}

func TestMoveFeedReordersSubscriptions(t *testing.T) {
	s := openTemp(t)
	var ids []int64
	for _, url := range []string{"https://a.example/feed", "https://b.example/feed", "https://c.example/feed", "https://d.example/feed"} {
		feed, err := s.AddFeed(Feed{URL: url})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, feed.ID)
	}

	if err := s.MoveFeed(ids[0], 2); err != nil {
		t.Fatal(err)
	}
	if got := feedIDOrder(s.Feeds()); !slices.Equal(got, []int64{ids[1], ids[2], ids[0], ids[3]}) {
		t.Fatalf("order after move = %v", got)
	}

	if err := s.MoveFeed(ids[3], 0); err != nil {
		t.Fatal(err)
	}
	if got := feedIDOrder(s.Feeds()); !slices.Equal(got, []int64{ids[3], ids[1], ids[2], ids[0]}) {
		t.Fatalf("order after move to front = %v", got)
	}

	if err := s.MoveFeed(ids[1], 99); err != nil {
		t.Fatal(err)
	}
	if got := feedIDOrder(s.Feeds()); !slices.Equal(got, []int64{ids[3], ids[2], ids[0], ids[1]}) {
		t.Fatalf("order after clamped move = %v", got)
	}

	if err := s.MoveFeed(9999, 0); !errors.Is(err, ErrNotFound) {
		t.Errorf("MoveFeed unknown id error = %v", err)
	}
}

func TestFeedOrderSurvivesReopen(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for _, url := range []string{"https://a.example/feed", "https://b.example/feed", "https://c.example/feed"} {
		feed, err := s.AddFeed(Feed{URL: url})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, feed.ID)
	}
	if err := s.MoveFeed(ids[2], 0); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	want := []int64{ids[2], ids[0], ids[1]}
	if got := feedIDOrder(reopened.Feeds()); !slices.Equal(got, want) {
		t.Fatalf("order after reopen = %v, want %v", got, want)
	}

	feed, err := reopened.AddFeed(Feed{URL: "https://d.example/feed"})
	if err != nil {
		t.Fatal(err)
	}
	if got := feedIDOrder(reopened.Feeds()); !slices.Equal(got, append(want, feed.ID)) {
		t.Fatalf("order after add = %v", got)
	}
}

func TestMigrationSeedsFeedOrder(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	var ids []int64
	for _, url := range []string{"https://a.example/feed", "https://b.example/feed", "https://c.example/feed"} {
		feed, err := s.AddFeed(Feed{URL: url})
		if err != nil {
			t.Fatal(err)
		}
		ids = append(ids, feed.ID)
	}

	// Simulate a v4 database: downgrade the version and drop the order key.
	err = s.db.Update(func(tx *bolt.Tx) error {
		settings := tx.Bucket([]byte(settingsBucket))
		if err := settings.Delete([]byte(feedOrderKey)); err != nil {
			return err
		}
		return put(settings, []byte(schemaVersionKey), 4)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	if got := feedIDOrder(reopened.Feeds()); !slices.Equal(got, ids) {
		t.Fatalf("seeded order = %v, want id order %v", got, ids)
	}
}
