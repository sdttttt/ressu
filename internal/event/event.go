// Package event defines the unified state event broadcast to every frontend
// and hosts the publisher shared by native and web views.
package event

import (
	"fmt"
	"strings"
	"time"

	appstatus "github.com/sdttttt/ressu/internal/status"
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
	// EventSelectionChanged carries the selected feed.
	EventSelectionChanged EventID = "selection.changed"
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

// Event is the single payload used by every Go-to-frontend broadcast. EID
// decides which fields consumers read; the other sections are null.
type Event struct {
	EID        EventID                  `json:"eid"`
	Revision   uint64                   `json:"revision"`
	OccurredAt time.Time                `json:"occurredAt"`
	Sync       *appstatus.SyncState     `json:"sync"`
	Feeds      *appstatus.FeedsSnapshot `json:"feeds"`
	Posts      *PostsChanged            `json:"posts"`
	Appearance *Appearance              `json:"appearance"`
	Selection  *Selection               `json:"selection"`
	Command    string                   `json:"command"`
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
	case EventSelectionChanged:
		if e.Selection == nil {
			return fmt.Errorf("event %q is missing selection", e.EID)
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
