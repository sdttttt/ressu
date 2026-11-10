package event

import (
	"sync"
	"testing"
	"time"

	appstatus "github.com/sdttttt/ressu/internal/status"
	"github.com/sdttttt/ressu/internal/store"
)

func TestEventValidate(t *testing.T) {
	syncState := &appstatus.SyncState{Active: true, Total: 1}
	feeds := &appstatus.FeedsSnapshot{}
	appearance := &Appearance{Theme: "system", FontSize: "medium"}
	prefs := &store.Prefs{}
	selection := &Selection{}
	posts := &PostsChanged{}

	tests := []struct {
		name  string
		event Event
		valid bool
	}{
		{"snapshot", Event{EID: EventStateSnapshot, Sync: syncState, Feeds: feeds, Appearance: appearance, Selection: selection}, true},
		{"sync started", Event{EID: EventSyncStarted, Sync: syncState}, true},
		{"sync progress", Event{EID: EventSyncProgress, Sync: syncState}, true},
		{"sync finished", Event{EID: EventSyncFinished, Sync: syncState}, true},
		{"feeds", Event{EID: EventFeedsChanged, Feeds: feeds}, true},
		{"posts", Event{EID: EventPostsChanged, Posts: posts}, true},
		{"appearance", Event{EID: EventAppearanceChanged, Appearance: appearance}, true},
		{"prefs", Event{EID: EventPrefsChanged, Prefs: prefs}, true},
		{"selection", Event{EID: EventSelectionChanged, Selection: selection}, true},
		{"command", Event{EID: EventUICommand, Command: "refresh"}, true},
		{"missing sync", Event{EID: EventSyncStarted}, false},
		{"missing feeds", Event{EID: EventFeedsChanged}, false},
		{"missing posts", Event{EID: EventPostsChanged}, false},
		{"missing appearance", Event{EID: EventAppearanceChanged}, false},
		{"missing prefs", Event{EID: EventPrefsChanged}, false},
		{"missing selection", Event{EID: EventSelectionChanged}, false},
		{"empty command", Event{EID: EventUICommand}, false},
		{"unknown", Event{EID: EventID("unknown")}, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.event.Validate()
			if test.valid && err != nil {
				t.Fatalf("Validate() = %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("Validate() = nil, want an error")
			}
		})
	}
}

func TestHubUpdatesModelAndStampsEvents(t *testing.T) {
	model := appstatus.New("1.2.3")
	var (
		mu      sync.Mutex
		emitted []Event
	)
	hub := newHub(model, func(e Event) error {
		mu.Lock()
		defer mu.Unlock()
		emitted = append(emitted, e)
		return nil
	})
	now := time.Date(2026, time.November, 9, 12, 0, 0, 0, time.FixedZone("CST", 8*60*60))
	hub.now = func() time.Time { return now }

	hub.FeedsChanged(appstatus.FeedsSnapshot{
		Feeds:  []store.Feed{{ID: 1, Title: "A"}},
		Unread: 3,
	})
	hub.SelectionChanged(1)
	hub.AppearanceChanged(store.ThemeSepia, store.FontSizeLarge)
	hub.PrefsChanged(store.Prefs{
		Theme:     store.ThemeSepia,
		FontSize:  store.FontSizeLarge,
		PaneSizes: []int{18, 30},
	})
	hub.SyncStarted(appstatus.SyncState{Active: true, Total: 2})
	hub.SyncProgress(appstatus.SyncState{Active: true, Done: 1, Total: 2, FeedID: 1, Feed: "A"})
	hub.PublishSnapshot()

	if len(emitted) != 7 {
		t.Fatalf("emitted = %d events", len(emitted))
	}
	for i, e := range emitted {
		if e.Revision != uint64(i+1) {
			t.Errorf("event %d revision = %d", i, e.Revision)
		}
		if !e.OccurredAt.Equal(now.UTC()) {
			t.Errorf("event %d time = %s, want %s", i, e.OccurredAt, now.UTC())
		}
	}

	prefsEvent := emitted[3]
	if prefsEvent.EID != EventPrefsChanged || prefsEvent.Prefs == nil ||
		prefsEvent.Prefs.Theme != store.ThemeSepia {
		t.Fatalf("prefs event = %+v", prefsEvent)
	}
	if len(prefsEvent.Prefs.PaneSizes) != 2 || prefsEvent.Prefs.PaneSizes[0] != 18 {
		t.Fatalf("prefs pane sizes = %v; want [18 30]", prefsEvent.Prefs.PaneSizes)
	}

	snapshot := emitted[6]
	if snapshot.EID != EventStateSnapshot || snapshot.Sync == nil || snapshot.Feeds == nil ||
		snapshot.Appearance == nil || snapshot.Selection == nil {
		t.Fatalf("snapshot = %+v", snapshot)
	}
	if snapshot.Sync.Done != 1 || snapshot.Sync.Total != 2 {
		t.Errorf("snapshot sync = %+v", snapshot.Sync)
	}
	if len(snapshot.Feeds.Feeds) != 1 || snapshot.Feeds.Unread != 3 {
		t.Errorf("snapshot feeds = %+v", snapshot.Feeds)
	}
	if snapshot.Appearance.Theme != store.ThemeSepia ||
		snapshot.Appearance.FontSize != store.FontSizeLarge {
		t.Errorf("snapshot appearance = %+v", snapshot.Appearance)
	}
	if snapshot.Selection.FeedID != 1 {
		t.Errorf("snapshot selection = %+v", snapshot.Selection)
	}
	if snapshot.Posts != nil || snapshot.Prefs != nil || snapshot.Command != "" {
		t.Errorf("snapshot carried transient state: %+v", snapshot)
	}

	modelSnapshot := model.Snapshot()
	if modelSnapshot.CurrentFeedID != 1 || modelSnapshot.FeedCount != 1 ||
		modelSnapshot.Unread != 3 || modelSnapshot.Sync.Done != 1 {
		t.Errorf("model snapshot = %+v", modelSnapshot)
	}
	if modelSnapshot.SidebarCollapsed || modelSnapshot.SidebarPercent != 18 {
		t.Errorf("model layout = %+v; want docked 18%%", modelSnapshot)
	}
}

func TestHubSerializesConcurrentPublishes(t *testing.T) {
	model := appstatus.New("1.0.0")
	var emitted []Event
	hub := newHub(model, func(e Event) error {
		emitted = append(emitted, e)
		return nil
	})

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hub.SyncProgress(appstatus.SyncState{Active: true, Total: 1})
		}()
	}
	wg.Wait()

	if len(emitted) != 20 {
		t.Fatalf("emitted = %d events", len(emitted))
	}
	for i, e := range emitted {
		if e.Revision != uint64(i+1) {
			t.Fatalf("event %d revision = %d", i, e.Revision)
		}
	}
}

func TestHubRejectsInvalidEvent(t *testing.T) {
	model := appstatus.New("1.0.0")
	emitted := 0
	hub := newHub(model, func(Event) error {
		emitted++
		return nil
	})

	hub.Publish(Event{EID: EventFeedsChanged})

	if emitted != 0 {
		t.Fatalf("emitted = %d invalid events", emitted)
	}
	if snapshot := model.Snapshot(); snapshot.FeedCount != 0 {
		t.Errorf("invalid event changed the model: %+v", snapshot)
	}
}
