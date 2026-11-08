package store

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

const (
	dbFileName       = "ressu.db"
	feedsBucket      = "feeds"
	postsBucket      = "posts"
	iconsBucket      = "icons"
	settingsBucket   = "settings"
	prefsKey         = "prefs"
	schemaVersionKey = "schemaVersion"
	iconDataKey      = "data"
	iconMetaKey      = "meta"
)

// currentSchemaVersion marks the newest data layout and cleanup behavior.
const currentSchemaVersion = 3

// maxIconBytes bounds the raw image kept in the local cache.
const maxIconBytes = 256 << 10

// Errors of the store.
var (
	ErrNotFound  = errors.New("not found")
	ErrDuplicate = errors.New("this feed is already subscribed")
)

// Store keeps the subscriptions and the preferences in memory, and the posts
// of every feed in a bbolt database. Every method is safe for concurrent use.
type Store struct {
	db *bolt.DB

	mu    sync.Mutex
	feeds []Feed
	prefs Prefs
}

// Open opens the database in dir, creating it with the default preferences
// when it is missing.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(filepath.Join(dir, dbFileName), 0o600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	s := &Store{db: db, prefs: DefaultPrefs()}
	if err := db.Update(func(tx *bolt.Tx) error {
		if err := s.init(tx); err != nil {
			return err
		}
		return migrate(tx)
	}); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// init creates the buckets and loads the subscriptions and preferences.
func (s *Store) init(tx *bolt.Tx) error {
	feeds, err := tx.CreateBucketIfNotExists([]byte(feedsBucket))
	if err != nil {
		return err
	}
	if _, err := tx.CreateBucketIfNotExists([]byte(postsBucket)); err != nil {
		return err
	}
	if _, err := tx.CreateBucketIfNotExists([]byte(iconsBucket)); err != nil {
		return err
	}
	settings, err := tx.CreateBucketIfNotExists([]byte(settingsBucket))
	if err != nil {
		return err
	}

	if err := feeds.ForEach(func(_, value []byte) error {
		var f Feed
		if err := json.Unmarshal(value, &f); err != nil {
			return err
		}
		s.feeds = append(s.feeds, f)
		return nil
	}); err != nil {
		return err
	}

	if _, err := get(settings, []byte(prefsKey), &s.prefs); err != nil {
		return err
	}
	s.prefs.Normalize()
	return put(settings, []byte(prefsKey), s.prefs)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// Dir returns the directory holding the database.
func (s *Store) Dir() string {
	return filepath.Dir(s.db.Path())
}

// feedKey encodes a feed id: big-endian bytes keep the keys in id order.
func feedKey(id int64) []byte {
	var key [8]byte
	binary.BigEndian.PutUint64(key[:], uint64(id))
	return key[:]
}

func put(b *bolt.Bucket, key []byte, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return b.Put(key, data)
}

func get(b *bolt.Bucket, key []byte, v any) (found bool, err error) {
	data := b.Get(key)
	if data == nil {
		return false, nil
	}
	return true, json.Unmarshal(data, v)
}

func (s *Store) putFeed(tx *bolt.Tx, f Feed) error {
	return put(tx.Bucket([]byte(feedsBucket)), feedKey(f.ID), f)
}

// updateUnread stores n as the unread count of a feed.
func (s *Store) updateUnread(tx *bolt.Tx, feedID int64, n int) error {
	i := s.feedIndex(feedID)
	if i < 0 {
		return ErrNotFound
	}
	f := s.feeds[i]
	f.Unread = n
	return s.putFeed(tx, f)
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
	if err := s.db.Update(func(tx *bolt.Tx) error { return s.putFeed(tx, f) }); err != nil {
		return Feed{}, err
	}
	s.feeds = append(s.feeds, f)
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
	if err := s.db.Update(func(tx *bolt.Tx) error { return s.putFeed(tx, f) }); err != nil {
		return err
	}
	s.feeds[i] = f
	return nil
}

// RemoveFeed deletes a subscription and its posts.
func (s *Store) RemoveFeed(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.feedIndex(id)
	if i < 0 {
		return ErrNotFound
	}
	if err := s.db.Update(func(tx *bolt.Tx) error {
		if err := deletePosts(tx, id); err != nil {
			return err
		}
		if err := deleteIcon(tx, id); err != nil {
			return err
		}
		return tx.Bucket([]byte(feedsBucket)).Delete(feedKey(id))
	}); err != nil {
		return err
	}
	s.feeds = slices.Delete(s.feeds, i, i+1)
	return nil
}

// Icon returns the cached icon of a feed.
func (s *Store) Icon(feedID int64) (Icon, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feedIndex(feedID) < 0 {
		return Icon{}, ErrNotFound
	}

	var icon Icon
	if err := s.db.View(func(tx *bolt.Tx) error {
		data, meta, err := readIcon(tx, feedID)
		if err != nil {
			return err
		}
		icon = Icon{Meta: meta, Data: slices.Clone(data)}
		return nil
	}); err != nil {
		return Icon{}, err
	}
	return icon, nil
}

// PutIcon replaces the cached icon of an existing feed.
func (s *Store) PutIcon(feedID int64, icon Icon) error {
	if len(icon.Data) > maxIconBytes {
		return errors.New("icon is too large")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feedIndex(feedID) < 0 {
		return ErrNotFound
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		feeds := tx.Bucket([]byte(iconsBucket))
		if feeds == nil {
			return errors.New("icons bucket is missing")
		}
		feed, err := feeds.CreateBucketIfNotExists(feedKey(feedID))
		if err != nil {
			return err
		}
		if err := feed.Put([]byte(iconDataKey), slices.Clone(icon.Data)); err != nil {
			return err
		}
		return put(feed, []byte(iconMetaKey), icon.Meta)
	})
}

// DeleteIcon removes the cached icon of a feed.
func (s *Store) DeleteIcon(feedID int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feedIndex(feedID) < 0 {
		return ErrNotFound
	}
	return s.db.Update(func(tx *bolt.Tx) error { return deleteIcon(tx, feedID) })
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
	if err := s.db.Update(func(tx *bolt.Tx) error {
		return put(tx.Bucket([]byte(settingsBucket)), []byte(prefsKey), p)
	}); err != nil {
		return err
	}
	s.prefs = p
	return nil
}

// Posts returns every post of a feed, newest first.
func (s *Store) Posts(feedID int64) ([]Post, error) {
	var posts []Post
	if err := s.db.View(func(tx *bolt.Tx) error {
		var err error
		posts, err = readPosts(tx, feedID)
		return err
	}); err != nil {
		return nil, err
	}
	slices.SortStableFunc(posts, byPublishedDesc)
	return posts, nil
}

// AllPosts returns every post of every feed, newest first.
func (s *Store) AllPosts() ([]Post, error) {
	var posts []Post
	if err := s.db.View(func(tx *bolt.Tx) error {
		return tx.Bucket([]byte(postsBucket)).ForEach(func(key, _ []byte) error {
			feedID := int64(binary.BigEndian.Uint64(key))
			feedPosts, err := readPosts(tx, feedID)
			if err != nil {
				return err
			}
			posts = append(posts, feedPosts...)
			return nil
		})
	}); err != nil {
		return nil, err
	}
	slices.SortStableFunc(posts, byPublishedDesc)
	return posts, nil
}

// Post returns one post of a feed.
func (s *Store) Post(feedID int64, guid string) (Post, error) {
	var post Post
	var found bool
	if err := s.db.View(func(tx *bolt.Tx) error {
		b := feedPosts(tx, feedID)
		if b == nil {
			return nil
		}
		var err error
		found, err = get(b, []byte(guid), &post)
		return err
	}); err != nil {
		return Post{}, err
	}
	if !found {
		return Post{}, ErrNotFound
	}
	return post, nil
}

// UpdatePost changes one post of a feed.
func (s *Store) UpdatePost(feedID int64, guid string, change func(*Post)) (Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.feedIndex(feedID)
	if i < 0 {
		return Post{}, ErrNotFound
	}

	var updated Post
	var unread int
	if err := s.db.Update(func(tx *bolt.Tx) error {
		b := feedPosts(tx, feedID)
		if b == nil {
			return ErrNotFound
		}
		var post Post
		found, err := get(b, []byte(guid), &post)
		if err != nil {
			return err
		}
		if !found {
			return ErrNotFound
		}

		read := post.Read
		change(&post)
		if err := put(b, []byte(guid), post); err != nil {
			return err
		}

		unread = s.feeds[i].Unread
		switch {
		case !read && post.Read:
			unread--
		case read && !post.Read:
			unread++
		}
		unread = max(unread, 0)
		updated = post
		return s.updateUnread(tx, feedID, unread)
	}); err != nil {
		return Post{}, err
	}
	s.feeds[i].Unread = unread
	return updated, nil
}

// MarkAllRead marks every post of a feed as read, and returns how many changed.
func (s *Store) MarkAllRead(feedID int64) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.feedIndex(feedID)
	if i < 0 {
		return 0, ErrNotFound
	}

	changed := 0
	if err := s.db.Update(func(tx *bolt.Tx) error {
		posts, err := readPosts(tx, feedID)
		if err != nil {
			return err
		}
		for j := range posts {
			if !posts[j].Read {
				posts[j].Read = true
				changed++
			}
		}
		if changed == 0 {
			return nil
		}
		if err := replacePosts(tx, feedID, posts); err != nil {
			return err
		}
		return s.updateUnread(tx, feedID, 0)
	}); err != nil {
		return 0, err
	}
	if changed == 0 {
		return 0, nil
	}
	s.feeds[i].Unread = 0
	return changed, nil
}

// MergePosts adds the incoming posts that are new, refreshes the ones that are
// already stored while keeping their read and starred flags, and saves the
// feed's posts. It returns how many posts were new.
func (s *Store) MergePosts(feedID int64, incoming []Post) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := s.feedIndex(feedID)
	if i < 0 {
		return 0, ErrNotFound
	}

	added := 0
	var unread int
	if err := s.db.Update(func(tx *bolt.Tx) error {
		posts, err := readPosts(tx, feedID)
		if err != nil {
			return err
		}

		index := make(map[string]int, len(posts))
		for j, p := range posts {
			index[p.GUID] = j
		}

		added = 0
		now := time.Now()
		for _, p := range incoming {
			p.FeedID = feedID
			if j, ok := index[p.GUID]; ok {
				kept := posts[j]
				p.Read = kept.Read
				p.Starred = kept.Starred
				p.FetchedAt = kept.FetchedAt
				posts[j] = p
				continue
			}
			p.FetchedAt = now
			posts = append(posts, p)
			index[p.GUID] = len(posts) - 1
			added++
		}

		slices.SortStableFunc(posts, byPublishedDesc)
		posts = s.trimPosts(posts)
		if err := replacePosts(tx, feedID, posts); err != nil {
			return err
		}
		unread = countUnread(posts)
		return s.updateUnread(tx, feedID, unread)
	}); err != nil {
		return 0, err
	}
	s.feeds[i].Unread = unread
	return added, nil
}

// feedPosts returns the bucket holding the posts of a feed, or nil.
func feedPosts(tx *bolt.Tx, feedID int64) *bolt.Bucket {
	return tx.Bucket([]byte(postsBucket)).Bucket(feedKey(feedID))
}

// readPosts decodes every post of a feed.
func readPosts(tx *bolt.Tx, feedID int64) ([]Post, error) {
	b := feedPosts(tx, feedID)
	if b == nil {
		return nil, nil
	}
	var posts []Post
	if err := b.ForEach(func(_, value []byte) error {
		var p Post
		if err := json.Unmarshal(value, &p); err != nil {
			return err
		}
		posts = append(posts, p)
		return nil
	}); err != nil {
		return nil, err
	}
	return posts, nil
}

// replacePosts writes the posts of a feed as a bucket keyed by GUID.
func replacePosts(tx *bolt.Tx, feedID int64, posts []Post) error {
	if err := deletePosts(tx, feedID); err != nil {
		return err
	}
	if len(posts) == 0 {
		return nil
	}
	b, err := tx.Bucket([]byte(postsBucket)).CreateBucket(feedKey(feedID))
	if err != nil {
		return err
	}
	for _, p := range posts {
		if err := put(b, []byte(p.GUID), p); err != nil {
			return err
		}
	}
	return nil
}

func deletePosts(tx *bolt.Tx, feedID int64) error {
	parent := tx.Bucket([]byte(postsBucket))
	key := feedKey(feedID)
	if parent.Bucket(key) == nil {
		return nil
	}
	return parent.DeleteBucket(key)
}

func (s *Store) trimPosts(posts []Post) []Post {
	limit := s.prefs.PostsPerFeed
	if limit <= 0 || len(posts) <= limit {
		return posts
	}
	return posts[:limit]
}

func countUnread(posts []Post) int {
	n := 0
	for _, p := range posts {
		if !p.Read {
			n++
		}
	}
	return n
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
