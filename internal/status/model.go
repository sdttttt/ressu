// Package status holds reader state that is shown outside the web page.
package status

import (
	"slices"
	"sync"
	"time"

	"github.com/sdttttt/ressu/internal/store"
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

// Snapshot is the state needed to draw the native status bar.
type Snapshot struct {
	Sync          SyncState
	FeedCount     int
	Unread        int
	CurrentFeed   *store.Feed
	Theme         string
	FontSize      string
	Version       string
	RelativeStart time.Time
}

// Model stores the service state mirrored by the native status bar.
type Model struct {
	mu      sync.RWMutex
	sync    SyncState
	feeds   []store.Feed
	unread  int
	current int64
	theme   string
	font    string
	version string
	changed func()
	now     func() time.Time
}

// New creates a status model showing the application version.
func New(version string) *Model {
	return &Model{
		theme:   store.ThemeSystem,
		font:    store.FontSizeMedium,
		version: version,
		now:     time.Now,
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

// Snapshot clones the state for one view frame.
func (m *Model) Snapshot() Snapshot {
	m.mu.RLock()
	feeds := slices.Clone(m.feeds)
	current := m.current
	snapshot := Snapshot{
		Sync:          m.sync,
		FeedCount:     len(feeds),
		Unread:        m.unread,
		Theme:         m.theme,
		FontSize:      m.font,
		Version:       m.version,
		RelativeStart: m.now(),
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
