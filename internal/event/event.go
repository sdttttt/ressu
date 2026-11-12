// Package event defines the unified state event broadcast to every frontend
// and hosts the publisher shared by native and web views.
package event

import (
	"fmt"
	"strings"
	"time"

	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

// EventID identifies how consumers should interpret an Event.
type EventID string

const (
	// EventStateSnapshot replaces every piece of broadcast state.
	EventStateSnapshot EventID = "state.snapshot"
	// EventSyncStarted announces the beginning of a refresh batch.
	EventSyncStarted EventID = "sync.started"
	// EventSyncProgress reports one completed feed in a refresh batch.
	EventSyncProgress EventID = "sync.progress"
	// EventSyncFinished announces the end of a refresh batch.
	EventSyncFinished EventID = "sync.finished"
	// EventFeedsChanged carries the complete subscription state.
	EventFeedsChanged EventID = "feeds.changed"
	// EventPostsChanged says which feed's posts changed.
	EventPostsChanged EventID = "posts.changed"
	// EventAppearanceChanged carries the theme and interface font size.
	EventAppearanceChanged EventID = "appearance.changed"
	// EventPrefsChanged carries the complete preferences.
	EventPrefsChanged EventID = "prefs.changed"
	// EventSelectionChanged carries the selected feed.
	EventSelectionChanged EventID = "selection.changed"
	// EventPostSelected carries the selected article.
	EventPostSelected EventID = "post.selected"
	// EventPostUpdated carries an article's read or starred state.
	EventPostUpdated EventID = "post.updated"
	// EventUICommand asks the page to run a native-menu command.
	EventUICommand EventID = "ui.command"
)

// PostsChanged says which feed's posts changed, and how many are new.
type PostsChanged struct {
	FeedID int64 `json:"feedId"`
	New    int   `json:"new"`
}

// Appearance is the reader's theme and interface font size.
type Appearance struct {
	Theme    string `json:"theme"`
	FontSize string `json:"fontSize"`
}

// Selection is the feed selected by the page. Zero selects all feeds.
type Selection struct {
	FeedID int64 `json:"feedId"`
}

// PostSelection is the article selected by the page.
type PostSelection = appstatus.PostSelection

// Event is the single payload used by every Go-to-frontend broadcast. EID
// decides which fields consumers read; the other sections are null.
type Event struct {
	EID           EventID                  `json:"eid"`
	Revision      uint64                   `json:"revision"`
	OccurredAt    time.Time                `json:"occurredAt"`
	Sync          *appstatus.SyncState     `json:"sync"`
	Feeds         *appstatus.FeedsSnapshot `json:"feeds"`
	Posts         *PostsChanged            `json:"posts"`
	Appearance    *Appearance              `json:"appearance"`
	Prefs         *store.Prefs             `json:"prefs"`
	Selection     *Selection               `json:"selection"`
	PostSelection *PostSelection           `json:"postSelection"`
	// Direction is where the selected post sits relative to the previously
	// selected one in the loaded list: 1 later, -1 earlier, 0 unknown.
	// Only EventPostSelected sets it.
	Direction int                `json:"direction"`
	Post      *store.PostSummary `json:"post"`
	Command   string             `json:"command"`
}

// Validate reports whether the event contains the fields its EID needs.
func (e Event) Validate() error {
	switch e.EID {
	case EventStateSnapshot:
		if e.Sync == nil || e.Feeds == nil || e.Appearance == nil || e.Selection == nil {
			return fmt.Errorf("event %q is missing snapshot state", e.EID)
		}
	case EventSyncStarted, EventSyncProgress, EventSyncFinished:
		if e.Sync == nil {
			return fmt.Errorf("event %q is missing sync state", e.EID)
		}
	case EventFeedsChanged:
		if e.Feeds == nil {
			return fmt.Errorf("event %q is missing feeds", e.EID)
		}
	case EventPostsChanged:
		if e.Posts == nil {
			return fmt.Errorf("event %q is missing post changes", e.EID)
		}
	case EventAppearanceChanged:
		if e.Appearance == nil {
			return fmt.Errorf("event %q is missing appearance", e.EID)
		}
	case EventPrefsChanged:
		if e.Prefs == nil {
			return fmt.Errorf("event %q is missing preferences", e.EID)
		}
	case EventSelectionChanged:
		if e.Selection == nil {
			return fmt.Errorf("event %q is missing selection", e.EID)
		}
	case EventPostSelected:
		if e.PostSelection == nil || e.PostSelection.GUID == "" ||
			e.Post == nil || e.Post.GUID == "" {
			return fmt.Errorf("event %q is missing a post selection", e.EID)
		}
	case EventPostUpdated:
		if e.Post == nil || e.Post.GUID == "" {
			return fmt.Errorf("event %q is missing a post update", e.EID)
		}
	case EventUICommand:
		if strings.TrimSpace(e.Command) == "" {
			return fmt.Errorf("event %q is missing a command", e.EID)
		}
	default:
		return fmt.Errorf("unknown event id %q", e.EID)
	}
	return nil
}
