// Package status holds reader state that is shown outside the web page.
package status

import (
	"slices"
	"sync"
	"time"

	"github.com/sdttttt/ressu/core/store"
)

// SyncState reports how a refresh is going.
type SyncState struct {
	Active bool   `json:"active"`
	Done   int    `json:"done"`
	Total  int    `json:"total"`
	FeedID int64  `json:"feedId"`
	Feed   string `json:"feed"`
	Error  string `json:"error,omitempty"`
}

// FeedsSnapshot is the subscription list with the total unread count.
type FeedsSnapshot struct {
	Feeds  []store.Feed `json:"feeds"`
	Unread int          `json:"unread"`
}

// PostSelection identifies the article selected by any reader surface.
type PostSelection struct {
	FeedID int64  `json:"feedId"`
	GUID   string `json:"guid"`
}

// Snapshot is the state needed to draw the native status bar.
type Snapshot struct {
	Sync             SyncState
	Feeds            []store.Feed
	FeedCount        int
	Unread           int
	CurrentFeedID    int64
	CurrentFeed      *store.Feed
	SidebarCollapsed bool
	SidebarPercent   float32
	ListPercent      float32
	Theme            string
	FontSize         string
	PostSort         string
	Version          string
	RelativeStart    time.Time
	PostsVersion     uint64
	PostVersion      uint64
	Post             store.PostSummary
	CurrentPost      PostSelection
	SelectedPost     store.PostSummary
}

// Model stores the service state mirrored by the native status bar.
type Model struct {
	mu               sync.RWMutex
	sync             SyncState
	feeds            []store.Feed
	unread           int
	current          int64
	currentPost      PostSelection
	sidebarCollapsed bool
	sidebarPercent   float32
	listPercent      float32
	postsVersion     uint64
	postVersion      uint64
	post             store.PostSummary
	selectedPost     store.PostSummary
	theme            string
	font             string
	postSort         string
	version          string
	changed          func()
	now              func() time.Time
}

// New creates a status model showing the application version.
func New(version string) *Model {
	return &Model{
		theme:    store.ThemeSystem,
		font:     store.FontSizeMedium,
		postSort: store.PostSortPublished,
		version:  version,
		now:      time.Now,
	}
}

// SetChanged installs the callback used to invalidate the native view.
func (m *Model) SetChanged(changed func()) {
	m.mu.Lock()
	m.changed = changed
	m.mu.Unlock()
}

// SetSync mirrors a refresh progress update.
func (m *Model) SetSync(state SyncState) {
	m.mu.Lock()
	m.sync = state
	changed := m.changed
	m.mu.Unlock()
	if changed != nil {
		changed()
	}
}

// SetFeeds mirrors the subscription list and total unread count.
func (m *Model) SetFeeds(snapshot FeedsSnapshot) {
	m.mu.Lock()
	m.feeds = slices.Clone(snapshot.Feeds)
	m.unread = snapshot.Unread
	changed := m.changed
	m.mu.Unlock()
	if changed != nil {
		changed()
	}
}

// SetCurrentFeed tells the status bar which feed the page selected.
func (m *Model) SetCurrentFeed(feedID int64) {
	m.mu.Lock()
	changed := m.current != feedID
	m.current = feedID
	if changed {
		m.currentPost = PostSelection{}
		m.selectedPost = store.PostSummary{}
	}
	callback := m.changed
	m.mu.Unlock()
	if changed && callback != nil {
		callback()
	}
}

// SetAppearance mirrors the reader's theme and interface font size.
func (m *Model) SetAppearance(theme, fontSize string) {
	m.mu.Lock()
	changed := m.theme != theme || m.font != fontSize
	m.theme = theme
	m.font = fontSize
	callback := m.changed
	m.mu.Unlock()
	if changed && callback != nil {
		callback()
	}
}

// SetPostSort mirrors the article list's preferred sort field.
func (m *Model) SetPostSort(sort string) {
	m.mu.Lock()
	changed := m.postSort != sort
	m.postSort = sort
	callback := m.changed
	m.mu.Unlock()
	if changed && callback != nil {
		callback()
	}
}

// SetLayout mirrors the sidebar's collapsed state and docked width.
func (m *Model) SetLayout(collapsed bool, sidebarPercent float32) {
	if sidebarPercent < 5 || sidebarPercent > 90 {
		sidebarPercent = 18
	}
	m.mu.Lock()
	changed := m.sidebarCollapsed != collapsed || m.sidebarPercent != sidebarPercent
	m.sidebarCollapsed = collapsed
	m.sidebarPercent = sidebarPercent
	callback := m.changed
	m.mu.Unlock()
	if changed && callback != nil {
		callback()
	}
}

// SetListPercent mirrors the article list's docked width percentage.
func (m *Model) SetListPercent(percent float32) {
	if percent < 18 || percent > 55 {
		percent = 30
	}
	m.mu.Lock()
	changed := m.listPercent != percent
	m.listPercent = percent
	callback := m.changed
	m.mu.Unlock()
	if changed && callback != nil {
		callback()
	}
}

// SetCurrentPost tells native UI which article the reader selected.
func (m *Model) SetCurrentPost(selection PostSelection, post store.PostSummary) {
	m.mu.Lock()
	changed := m.currentPost != selection || m.selectedPost != post
	m.currentPost = selection
	m.selectedPost = post
	callback := m.changed
	m.mu.Unlock()
	if changed && callback != nil {
		callback()
	}
}

// SetPostsChanged records that one feed's article list is stale.
func (m *Model) SetPostsChanged() {
	m.mu.Lock()
	m.postsVersion++
	callback := m.changed
	m.mu.Unlock()
	if callback != nil {
		callback()
	}
}

// SetPostUpdated mirrors one article's read or starred state.
func (m *Model) SetPostUpdated(post store.PostSummary) {
	m.mu.Lock()
	m.postVersion++
	m.post = post
	if m.currentPost.FeedID == post.FeedID && m.currentPost.GUID == post.GUID {
		m.selectedPost = post
	}
	callback := m.changed
	m.mu.Unlock()
	if callback != nil {
		callback()
	}
}

// Snapshot clones the state for one view frame.
func (m *Model) Snapshot() Snapshot {
	m.mu.RLock()
	feeds := slices.Clone(m.feeds)
	current := m.current
	snapshot := Snapshot{
		Sync:             m.sync,
		Feeds:            feeds,
		FeedCount:        len(feeds),
		Unread:           m.unread,
		CurrentFeedID:    current,
		SidebarCollapsed: m.sidebarCollapsed,
		SidebarPercent:   m.sidebarPercent,
		ListPercent:      m.listPercent,
		Theme:            m.theme,
		FontSize:         m.font,
		PostSort:         m.postSort,
		Version:          m.version,
		RelativeStart:    m.now(),
		PostsVersion:     m.postsVersion,
		PostVersion:      m.postVersion,
		Post:             m.post,
		CurrentPost:      m.currentPost,
		SelectedPost:     m.selectedPost,
	}
	m.mu.RUnlock()

	if current != 0 {
		index := slices.IndexFunc(feeds, func(feed store.Feed) bool {
			return feed.ID == current
		})
		if index >= 0 {
			snapshot.CurrentFeed = &feeds[index]
		}
	}
	return snapshot
}
