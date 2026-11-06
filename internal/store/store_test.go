package store

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

var base = time.Date(2024, 10, 1, 12, 0, 0, 0, time.UTC)

func openTemp(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestOpenCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"feeds.json", "settings.json", "posts"} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
	if got := s.Prefs().PostsPerFeed; got != DefaultPostsPerFeed {
		t.Errorf("posts per feed = %d", got)
	}
	if len(s.Feeds()) != 0 {
		t.Error("a new store should have no feeds")
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
	if _, err := os.Stat(filepath.Join(s.Dir(), "posts", "1.json")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("posts file should be gone: %v", err)
	}
	if len(s.Feeds()) != 0 {
		t.Error("feeds should be empty")
	}
}

func TestPrefsSurviveReopen(t *testing.T) {
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
	prefs := s.Prefs()
	prefs.Theme = ThemeDark
	prefs.PostsPerFeed = 100
	prefs.PaneSizes = []int{40, 55}
	if err := s.SetPrefs(prefs); err != nil {
		t.Fatal(err)
	}

	reopened, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	got := reopened.Prefs()
	if got.Theme != ThemeDark || got.PostsPerFeed != 100 {
		t.Errorf("prefs = %+v", got)
	}
	if len(got.PaneSizes) != 2 || got.PaneSizes[0] != 40 {
		t.Errorf("pane sizes = %v", got.PaneSizes)
	}

	feeds := reopened.Feeds()
	if len(feeds) != 1 || feeds[0].Title != "A" {
		t.Errorf("feeds = %+v", feeds)
	}
	posts, err := reopened.Posts(feeds[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(posts) != 1 || posts[0].Title != "One" {
		t.Errorf("posts = %+v", posts)
	}
}

func TestPrefsNormalize(t *testing.T) {
	p := Prefs{Theme: "sepia", OpenLinks: "carrier-pigeon", PostsPerFeed: 3, PaneSizes: []int{10}}
	p.Normalize()

	if p.Theme != ThemeSystem {
		t.Errorf("theme = %q", p.Theme)
	}
	if p.OpenLinks != OpenLinksBrowser {
		t.Errorf("open links = %q", p.OpenLinks)
	}
	if p.PostsPerFeed != DefaultPostsPerFeed {
		t.Errorf("posts per feed = %d", p.PostsPerFeed)
	}
	if len(p.PaneSizes) != 2 {
		t.Errorf("pane sizes = %v", p.PaneSizes)
	}
}

func TestWriteJSONReplacesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.json")

	if err := writeJSON(path, map[string]int{"a": 1}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(path, map[string]int{"b": 2}); err != nil {
		t.Fatal(err)
	}

	got, found, err := readJSON[map[string]int](path)
	if err != nil || !found {
		t.Fatalf("read: found=%v err=%v", found, err)
	}
	if _, ok := got["a"]; ok {
		t.Error("the old content should be gone")
	}
	if got["b"] != 2 {
		t.Errorf("data = %v", got)
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("temporary files left behind: %v", entries)
	}
}
