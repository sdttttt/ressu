package store

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"
)

const (
	feedsFileName    = "feeds.json"
	settingsFileName = "settings.json"
	postsDirName     = "posts"
	schemaVersion    = 1
)

// Errors of the store.
var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("this feed is already subscribed")
)

type feedsFile struct {
	Version int    `json:"version"`
	Feeds   []Feed `json:"feeds"`
}

type postsFile struct {
	Version int    `json:"version"`
	Posts   []Post `json:"posts"`
}

type settingsFile struct {
	Version int   `json:"version"`
	Prefs   Prefs `json:"prefs"`
}

// Store holds the subscriptions and preferences, and the posts of each feed
// once they are needed. Every method is safe for concurrent use.
type Store struct {
	dir string

	mu    sync.Mutex
	feeds []Feed
	prefs Prefs
	posts map[int64][]Post
}

// Open loads the store in dir, creating the directory and its default settings
// when they are missing.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(filepath.Join(dir, postsDirName), 0o755); err != nil {
		return nil, err
	}

	s := &Store{dir: dir, prefs: DefaultPrefs(), posts: map[int64][]Post{}}

	feeds, hasFeeds, err := readJSON[feedsFile](s.feedsPath())
	if err != nil {
		return nil, err
	}
	if hasFeeds {
		s.feeds = feeds.Feeds
	}

	settings, hasSettings, err := readJSON[settingsFile](s.settingsPath())
	if err != nil {
		return nil, err
	}
	if hasSettings {
		s.prefs = settings.Prefs
	}
	s.prefs.Normalize()

	if !hasFeeds {
		if err := s.saveFeeds(); err != nil {
			return nil, err
		}
	}
	if !hasSettings {
		if err := s.saveSettings(); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Dir returns the directory holding the store's files.
func (s *Store) Dir() string {
	return s.dir
}

func (s *Store) feedsPath() string {
	return filepath.Join(s.dir, feedsFileName)
}

func (s *Store) settingsPath() string {
	return filepath.Join(s.dir, settingsFileName)
}

func (s *Store) postsPath(feedID int64) string {
	return filepath.Join(s.dir, postsDirName, strconv.FormatInt(feedID, 10)+".json")
}

func (s *Store) saveFeeds() error {
	return writeJSON(s.feedsPath(), feedsFile{Version: schemaVersion, Feeds: s.feeds})
}

func (s *Store) saveSettings() error {
	return writeJSON(s.settingsPath(), settingsFile{Version: schemaVersion, Prefs: s.prefs})
}

func (s *Store) feedIndex(id int64) int {
	return slices.IndexFunc(s.feeds, func(f Feed) bool { return f.ID == id })
}

func (s *Store) nextFeedID() int64 {
	var maxID int64
	for _, f := range s.feeds {
		maxID = max(maxID, f.ID)
	}
	return maxID + 1
}

// Feeds returns the subscriptions.
func (s *Store) Feeds() []Feed {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.feeds)
}

// Unread returns how many posts are unread across every feed.
func (s *Store) Unread() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	for _, f := range s.feeds {
		total += f.Unread
	}
	return total
}

// Feed returns one subscription.
func (s *Store) Feed(id int64) (Feed, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i := s.feedIndex(id)
	if i < 0 {
		return Feed{}, ErrNotFound
	}
	return s.feeds[i], nil
}

// AddFeed stores a new subscription and returns it with its ID.
func (s *Store) AddFeed(f Feed) (Feed, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, exist := range s.feeds {
		if exist.URL == f.URL {
			return Feed{}, ErrDuplicate
		}
	}
	f.ID = s.nextFeedID()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = time.Now()
	}
	s.feeds = append(s.feeds, f)
	if err := s.saveFeeds(); err != nil {
		return Feed{}, err
	}
	return f, nil
}

// UpdateFeed replaces the metadata of a subscription.
func (s *Store) UpdateFeed(f Feed) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.feedIndex(f.ID)
	if i < 0 {
		return ErrNotFound
	}
	s.feeds[i] = f
	return s.saveFeeds()
}

// RemoveFeed deletes a subscription and its posts.
func (s *Store) RemoveFeed(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.feedIndex(id)
	if i < 0 {
		return ErrNotFound
	}
	s.feeds = slices.Delete(s.feeds, i, i+1)
	delete(s.posts, id)
	if err := os.Remove(s.postsPath(id)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return s.saveFeeds()
}

// Prefs returns the preferences.
func (s *Store) Prefs() Prefs {
	s.mu.Lock()
	defer s.mu.Unlock()
	p := s.prefs
	p.PaneSizes = slices.Clone(s.prefs.PaneSizes)
	return p
}

// SetPrefs stores the preferences.
func (s *Store) SetPrefs(p Prefs) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p.Normalize()
	s.prefs = p
	return s.saveSettings()
}

// Posts returns every post of a feed, newest first.
func (s *Store) Posts(feedID int64) ([]Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.postsLocked(feedID)
}

// Post returns one post of a feed.
func (s *Store) Post(feedID int64, guid string) (Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	posts, err := s.postsLocked(feedID)
	if err != nil {
		return Post{}, err
	}
	i := slices.IndexFunc(posts, func(p Post) bool { return p.GUID == guid })
	if i < 0 {
		return Post{}, ErrNotFound
	}
	return posts[i], nil
}

// UpdatePost changes one post of a feed and saves the feed's posts.
func (s *Store) UpdatePost(feedID int64, guid string, change func(*Post)) (Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	posts, err := s.postsLocked(feedID)
	if err != nil {
		return Post{}, err
	}
	i := slices.IndexFunc(posts, func(p Post) bool { return p.GUID == guid })
	if i < 0 {
		return Post{}, ErrNotFound
	}
	change(&posts[i])
	if err := s.savePostsLocked(feedID, posts); err != nil {
		return Post{}, err
	}
	return posts[i], nil
}

// MarkAllRead marks every post of a feed as read, and returns how many changed.
func (s *Store) MarkAllRead(feedID int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	posts, err := s.postsLocked(feedID)
	if err != nil {
		return 0, err
	}
	changed := 0
	for i := range posts {
		if !posts[i].Read {
			posts[i].Read = true
			changed++
		}
	}
	if changed == 0 {
		return 0, nil
	}
	if err := s.savePostsLocked(feedID, posts); err != nil {
		return 0, err
	}
	return changed, nil
}

// MergePosts adds the incoming posts that are new, refreshes the ones that are
// already stored while keeping their read and starred flags, and saves the
// feed's posts. It returns how many posts were new.
func (s *Store) MergePosts(feedID int64, incoming []Post) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	posts, err := s.postsLocked(feedID)
	if err != nil {
		return 0, err
	}

	index := make(map[string]int, len(posts))
	for i, p := range posts {
		index[p.GUID] = i
	}

	added := 0
	now := time.Now()
	for _, p := range incoming {
		p.FeedID = feedID
		if i, ok := index[p.GUID]; ok {
			kept := posts[i]
			p.Read = kept.Read
			p.Starred = kept.Starred
			p.FetchedAt = kept.FetchedAt
			posts[i] = p
			continue
		}
		p.FetchedAt = now
		posts = append(posts, p)
		added++
	}

	slices.SortStableFunc(posts, byPublishedDesc)
	posts = s.trimPosts(posts)
	if err := s.savePostsLocked(feedID, posts); err != nil {
		return 0, err
	}
	return added, nil
}

func (s *Store) postsLocked(feedID int64) ([]Post, error) {
	if cached, ok := s.posts[feedID]; ok {
		return slices.Clone(cached), nil
	}

	file, found, err := readJSON[postsFile](s.postsPath(feedID))
	if err != nil {
		return nil, err
	}
	posts := file.Posts
	if !found {
		posts = nil
	}
	slices.SortStableFunc(posts, byPublishedDesc)
	s.posts[feedID] = posts
	return slices.Clone(posts), nil
}

// savePostsLocked writes a feed's posts, refreshes its unread count and saves
// the subscriptions.
func (s *Store) savePostsLocked(feedID int64, posts []Post) error {
	if err := writeJSON(s.postsPath(feedID), postsFile{Version: schemaVersion, Posts: posts}); err != nil {
		return err
	}
	s.posts[feedID] = posts

	i := s.feedIndex(feedID)
	if i < 0 {
		return ErrNotFound
	}
	unread := 0
	for _, p := range posts {
		if !p.Read {
			unread++
		}
	}
	s.feeds[i].Unread = unread
	return s.saveFeeds()
}

func (s *Store) trimPosts(posts []Post) []Post {
	limit := s.prefs.PostsPerFeed
	if limit <= 0 || len(posts) <= limit {
		return posts
	}
	return posts[:limit]
}

func byPublishedDesc(a, b Post) int {
	switch {
	case a.PublishedAt.After(b.PublishedAt):
		return -1
	case a.PublishedAt.Before(b.PublishedAt):
		return 1
	default:
		return 0
	}
}
