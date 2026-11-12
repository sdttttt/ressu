package event

import (
	"log/slog"
	"sync"
	"time"

	"github.com/egoist/mygo"

	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

// AppEvent is the only event channel the Go side broadcasts to pages.
var AppEvent = mygo.NewEvent[Event]("app:event")

// Hub serializes broadcasts and mirrors each event into the native UI model.
type Hub struct {
	mu       sync.Mutex
	revision uint64
	model    *appstatus.Model
	now      func() time.Time
	sink     func(Event) error
}

// NewHub creates a publisher that updates model before broadcasting to pages.
func NewHub(model *appstatus.Model) *Hub {
	return newHub(model, AppEvent.Broadcast)
}

func newHub(model *appstatus.Model, sink func(Event) error) *Hub {
	return &Hub{
		model: model,
		now:   time.Now,
		sink:  sink,
	}
}

// SyncStarted broadcasts the beginning of a refresh batch.
func (h *Hub) SyncStarted(state appstatus.SyncState) {
	h.Publish(Event{EID: EventSyncStarted, Sync: &state})
}

// SyncProgress broadcasts one completed feed in a refresh batch.
func (h *Hub) SyncProgress(state appstatus.SyncState) {
	h.Publish(Event{EID: EventSyncProgress, Sync: &state})
}

// SyncFinished broadcasts the end of a refresh batch.
func (h *Hub) SyncFinished(state appstatus.SyncState) {
	h.Publish(Event{EID: EventSyncFinished, Sync: &state})
}

// FeedsChanged broadcasts the complete subscription state.
func (h *Hub) FeedsChanged(snapshot appstatus.FeedsSnapshot) {
	h.Publish(Event{EID: EventFeedsChanged, Feeds: &snapshot})
}

// PostsChanged broadcasts which feed's posts changed.
func (h *Hub) PostsChanged(change PostsChanged) {
	h.Publish(Event{EID: EventPostsChanged, Posts: &change})
}

// AppearanceChanged broadcasts the current theme and interface font size.
func (h *Hub) AppearanceChanged(theme, fontSize string) {
	h.Publish(Event{
		EID:        EventAppearanceChanged,
		Appearance: &Appearance{Theme: theme, FontSize: fontSize},
	})
}

// PrefsChanged broadcasts the complete preferences.
func (h *Hub) PrefsChanged(prefs store.Prefs) {
	h.Publish(Event{EID: EventPrefsChanged, Prefs: &prefs})
}

// SelectionChanged broadcasts the feed selected by the page.
func (h *Hub) SelectionChanged(feedID int64) {
	h.Publish(Event{
		EID:       EventSelectionChanged,
		Selection: &Selection{FeedID: feedID},
	})
}

// PostSelected broadcasts the article selected by a reader surface. The
// direction tells surfaces how the new post relates to the previous
// selection in the list: 1 later, -1 earlier, 0 unknown.
func (h *Hub) PostSelected(selection appstatus.PostSelection, post store.PostSummary, direction int) {
	h.Publish(Event{
		EID:           EventPostSelected,
		PostSelection: &selection,
		Direction:     direction,
		Post:          &post,
	})
}

// PostUpdated broadcasts an article's read or starred state.
func (h *Hub) PostUpdated(post store.PostSummary) {
	h.Publish(Event{EID: EventPostUpdated, Post: &post})
}

// UICommand asks every page to run a native-menu command.
func (h *Hub) UICommand(command string) {
	h.Publish(Event{EID: EventUICommand, Command: command})
}

// SelectedPost reports the article a reader surface selected last.
func (h *Hub) SelectedPost() (appstatus.PostSelection, store.PostSummary) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.model == nil {
		return appstatus.PostSelection{}, store.PostSummary{}
	}
	snapshot := h.model.Snapshot()
	return snapshot.CurrentPost, snapshot.SelectedPost
}

// PublishSnapshot broadcasts the complete current state.
func (h *Hub) PublishSnapshot() {
	if h == nil || h.model == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.publishLocked(h.snapshotLocked())
}

// Publish validates, stamps, applies, and broadcasts one event.
func (h *Hub) Publish(e Event) {
	if h == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	h.publishLocked(e)
}

func (h *Hub) publishLocked(e Event) {
	if err := e.Validate(); err != nil {
		slog.Error("event validation failed", "operation", "event.publish", "error", err)
		return
	}

	h.revision++
	e.Revision = h.revision
	e.OccurredAt = h.now().UTC()
	h.applyLocked(e)

	if h.sink == nil {
		return
	}
	if err := h.sink(e); err != nil {
		slog.Warn("event broadcast failed",
			"operation", "event.publish", "eid", e.EID, "error", err)
	}
}

func (h *Hub) applyLocked(e Event) {
	if h.model == nil {
		return
	}

	switch e.EID {
	case EventSyncStarted, EventSyncProgress, EventSyncFinished:
		h.model.SetSync(*e.Sync)
	case EventFeedsChanged:
		h.model.SetFeeds(*e.Feeds)
	case EventAppearanceChanged:
		h.model.SetAppearance(e.Appearance.Theme, e.Appearance.FontSize)
	case EventPrefsChanged:
		sidebarPercent := float32(18)
		if len(e.Prefs.PaneSizes) == 2 {
			sidebarPercent = float32(e.Prefs.PaneSizes[0])
		}
		h.model.SetLayout(e.Prefs.SidebarCollapsed, sidebarPercent)
		listPercent := float32(30)
		if len(e.Prefs.PaneSizes) == 2 {
			listPercent = float32(e.Prefs.PaneSizes[1])
		}
		h.model.SetListPercent(listPercent)
		h.model.SetPostSort(e.Prefs.PostSort)
	case EventSelectionChanged:
		h.model.SetCurrentFeed(e.Selection.FeedID)
	case EventPostsChanged:
		h.model.SetPostsChanged()
	case EventPostSelected:
		h.model.SetCurrentPost(*e.PostSelection, *e.Post)
	case EventPostUpdated:
		h.model.SetPostUpdated(*e.Post)
	}
}

func (h *Hub) snapshotLocked() Event {
	snapshot := h.model.Snapshot()
	return Event{
		EID:  EventStateSnapshot,
		Sync: &snapshot.Sync,
		Feeds: &appstatus.FeedsSnapshot{
			Feeds:  snapshot.Feeds,
			Unread: snapshot.Unread,
		},
		Appearance: &Appearance{
			Theme:    snapshot.Theme,
			FontSize: snapshot.FontSize,
		},
		Selection: &Selection{FeedID: snapshot.CurrentFeedID},
	}
}
