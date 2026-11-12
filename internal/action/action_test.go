package action

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	appevent "github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/core/service"
	appstatus "github.com/sdttttt/ressu/core/status"
	"github.com/sdttttt/ressu/core/store"
)

type fakeUI struct {
	openFiles             []string
	savePath              string
	openedPath            string
	externalLink          string
	windowTitle           string
	windowLink            string
	theme                 string
	snapshotData          []byte
	snapshotTitle         string
	snapshotParent        WindowHandle
	snapshotCallbacks     SnapshotViewerCallbacks
	snapshotCaptureParent WindowHandle
	snapshotCaptureClosed WindowHandle
	notifications         []Notification
	addFeedOpened         bool
	helpShown             bool
	feedDelta             int
	confirmation          *ConfirmationRequest
	confirmed             bool
	confirmCalled         bool
}

func (f *fakeUI) SelectOPMLImport(WindowHandle) ([]string, error) {
	return f.openFiles, nil
}

func (f *fakeUI) SelectOPMLExport(WindowHandle, string) (string, error) {
	return f.savePath, nil
}

func (f *fakeUI) SelectSnapshotSavePath(WindowHandle, string) (string, error) {
	return f.savePath, nil
}

func (f *fakeUI) OpenPath(path string) error {
	f.openedPath = path
	return nil
}

func (f *fakeUI) OpenExternal(link string) error {
	f.externalLink = link
	return nil
}

func (f *fakeUI) OpenLinkWindow(title, link string) {
	f.windowTitle = title
	f.windowLink = link
}

func (f *fakeUI) OpenImageViewer(WindowHandle, []byte, string, string) error {
	return nil
}

func (f *fakeUI) OpenSnapshotViewer(
	parent WindowHandle,
	data []byte,
	title string,
	callbacks SnapshotViewerCallbacks,
) error {
	f.snapshotData = data
	f.snapshotTitle = title
	f.snapshotCallbacks = callbacks
	f.snapshotParent = parent
	return nil
}

func (f *fakeUI) OpenSettings(WindowHandle, SettingsCallbacks) {}

func (f *fakeUI) OpenAddFeed(WindowHandle, AddFeedCallbacks) {
	f.addFeedOpened = true
}

func (f *fakeUI) OpenSnapshotCapture(parent WindowHandle) {
	f.snapshotCaptureParent = parent
}

func (f *fakeUI) CloseSnapshotCapture(window WindowHandle) {
	f.snapshotCaptureClosed = window
}

func (f *fakeUI) Confirm(
	_ context.Context,
	_ WindowHandle,
	request ConfirmationRequest,
) (bool, error) {
	f.confirmation = &request
	f.confirmCalled = true
	return f.confirmed, nil
}

func (f *fakeUI) CapturePage(context.Context) ([]byte, error) {
	return []byte("png"), nil
}

func (f *fakeUI) WriteImage([]byte) error {
	return nil
}

func (f *fakeUI) ApplyTheme(theme string) {
	f.theme = theme
}

func (f *fakeUI) ShowNotification(notification Notification) {
	f.notifications = append(f.notifications, notification)
}

func (f *fakeUI) SelectNeighbourPost(int, bool) {}

func (f *fakeUI) SelectNeighbourFeed(delta int) {
	f.feedDelta = delta
}

func (f *fakeUI) ShowHelp() {
	f.helpShown = true
}

func newTestActions(t *testing.T) (*AppActions, *fakeUI) {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})
	ui := &fakeUI{}
	settings := service.NewSettings(st)
	return New(nil, nil, settings, nil, nil, ui), ui
}

func TestOpenLinkRejectsUnsafeURL(t *testing.T) {
	actions, ui := newTestActions(t)

	for _, link := range []string{
		"javascript:alert(1)",
		"file:///tmp/secret",
		"ftp://example.com/file",
		"example.com/post",
	} {
		err := actions.OpenLink(link)
		if !errors.Is(err, ErrInvalidLink) {
			t.Errorf("OpenLink(%q) = %v; want ErrInvalidLink", link, err)
		}
	}
	if ui.externalLink != "" || ui.windowLink != "" {
		t.Fatalf("unsafe links reached the UI: external=%q window=%q", ui.externalLink, ui.windowLink)
	}
}

func TestOpenAddFeedUsesNativeUI(t *testing.T) {
	actions, ui := newTestActions(t)

	actions.OpenAddFeed(WindowHandle{})

	if !ui.addFeedOpened {
		t.Error("OpenAddFeed did not reach the native UI")
	}
}

func TestConfirmRemoveFeedDeletesAfterConfirmation(t *testing.T) {
	actions, ui, st := newRemovalTestActions(t)
	feed, err := st.AddFeed(store.Feed{URL: "https://example.com/feed.xml", Title: "测试源"})
	if err != nil {
		t.Fatal(err)
	}
	ui.confirmed = true

	if err := actions.ConfirmRemoveFeed(context.Background(), WindowHandle{}, feed.ID); err != nil {
		t.Fatal(err)
	}
	if !ui.confirmCalled {
		t.Fatal("RemoveFeed bypassed confirmation")
	}
	if got := ui.confirmation.Title; got != "取消订阅「测试源」？" {
		t.Errorf("confirmation title = %q; want %q", got, "取消订阅「测试源」？")
	}
	if !ui.confirmation.Destructive {
		t.Error("destructive confirmation did not reach the UI")
	}
	if _, err := st.Feed(feed.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("st.Feed(%d) = %v; want ErrNotFound", feed.ID, err)
	}
}

func TestConfirmRemoveFeedKeepsFeedWhenCanceled(t *testing.T) {
	actions, ui, st := newRemovalTestActions(t)
	feed, err := st.AddFeed(store.Feed{URL: "https://example.com/feed.xml", Title: "测试源"})
	if err != nil {
		t.Fatal(err)
	}

	if err := actions.ConfirmRemoveFeed(context.Background(), WindowHandle{}, feed.ID); err != nil {
		t.Fatal(err)
	}
	if !ui.confirmCalled {
		t.Fatal("CancelRemoveFeed did not reach the UI")
	}
	if _, err := st.Feed(feed.ID); err != nil {
		t.Fatalf("canceled removal removed feed: %v", err)
	}
}

func TestConfirmRemoveFeedRejectsMissingFeed(t *testing.T) {
	actions, ui, _ := newRemovalTestActions(t)

	if err := actions.ConfirmRemoveFeed(context.Background(), WindowHandle{}, 1); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("ConfirmRemoveFeed(1) = %v; want ErrNotFound", err)
	}
	if ui.confirmCalled {
		t.Error("a missing feed showed a confirmation")
	}
}

func newRemovalTestActions(t *testing.T) (*AppActions, *fakeUI, *store.Store) {
	t.Helper()
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})
	ui := &fakeUI{}
	feeds := service.NewFeeds(st)
	return New(feeds, nil, nil, nil, nil, ui), ui, st
}

func TestOpenLinkUsesConfiguredSurface(t *testing.T) {
	actions, ui := newTestActions(t)

	if err := actions.OpenLink("https://example.com/post"); err != nil {
		t.Fatal(err)
	}
	if ui.externalLink != "https://example.com/post" || ui.windowLink != "" {
		t.Fatalf("browser open = external %q, window %q", ui.externalLink, ui.windowLink)
	}

	prefs := store.DefaultPrefs()
	prefs.OpenLinks = store.OpenLinksWindow
	if _, err := actions.UpdatePreferences(prefs); err != nil {
		t.Fatal(err)
	}
	if err := actions.OpenLink("https://example.com/post"); err != nil {
		t.Fatal(err)
	}
	if ui.windowTitle != "https://example.com/post" || ui.windowLink != "https://example.com/post" {
		t.Fatalf("window open = title %q, link %q", ui.windowTitle, ui.windowLink)
	}
}

func TestUpdateSettingsDraftPreservesLayout(t *testing.T) {
	actions, ui := newTestActions(t)
	prefs := store.DefaultPrefs()
	prefs.PaneSizes = []int{27, 41}
	prefs.SidebarCollapsed = true
	if _, err := actions.UpdatePreferences(prefs); err != nil {
		t.Fatal(err)
	}

	draft := prefs
	draft.Theme = store.ThemeSepia
	draft.PaneSizes = []int{12, 18}
	draft.SidebarCollapsed = false
	stored, err := actions.UpdateSettingsDraft(draft)
	if err != nil {
		t.Fatal(err)
	}

	if stored.Theme != store.ThemeSepia {
		t.Errorf("theme = %q; want sepia", stored.Theme)
	}
	if len(stored.PaneSizes) != 2 || stored.PaneSizes[0] != 27 || stored.PaneSizes[1] != 41 {
		t.Errorf("pane sizes = %v; want [27 41]", stored.PaneSizes)
	}
	if !stored.SidebarCollapsed {
		t.Error("sidebar collapsed state changed")
	}
	if ui.theme != store.ThemeSepia {
		t.Errorf("native theme = %q; want sepia", ui.theme)
	}
}

func TestNotifyForwardsThroughNativeUI(t *testing.T) {
	actions, ui := newTestActions(t)

	actions.Notify("warning", "刷新失败")

	if len(ui.notifications) != 1 {
		t.Fatalf("notifications = %d; want 1", len(ui.notifications))
	}
	if got := ui.notifications[0]; got.Type != "warning" || got.Title != "刷新失败" {
		t.Errorf("notification = %#v; want warning/刷新失败", got)
	}
}

func TestPageShortcutsForwardThroughNativeUI(t *testing.T) {
	actions, ui := newTestActions(t)

	actions.ShowHelp()
	actions.SelectNeighbourFeed(-1)

	if !ui.helpShown {
		t.Error("help shortcut was not forwarded")
	}
	if ui.feedDelta != -1 {
		t.Errorf("feed delta = %d; want -1", ui.feedDelta)
	}
}

func TestSetSidebarPercentClampsPreference(t *testing.T) {
	actions, _ := newTestActions(t)

	actions.SetSidebarPercent(4)
	if got := actions.Preferences().PaneSizes[0]; got != 5 {
		t.Errorf("sidebar percent = %d; want 5", got)
	}

	actions.SetSidebarPercent(95)
	if got := actions.Preferences().PaneSizes[0]; got != 90 {
		t.Errorf("sidebar percent = %d; want 90", got)
	}
}

func TestAddAndSelectFeedMirrorsSelection(t *testing.T) {
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>测试源</title>
<link>https://example.com/</link><description>测试</description>
<item><title>第一篇</title><link>https://example.com/1</link>
<guid>guid-1</guid><pubDate>Wed, 02 Oct 2024 12:00:00 GMT</pubDate></item>
</channel></rss>`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = io.WriteString(w, feedXML)
	}))
	t.Cleanup(server.Close)

	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})

	model := appstatus.New("")
	hub := appevent.NewHub(model, nil)
	actions := New(
		service.NewFeedsWithHub(st, hub),
		service.NewPosts(st),
		service.NewSettings(st),
		service.NewStatus(hub, nil),
		nil,
		&fakeUI{},
	)

	feed, err := actions.AddAndSelectFeed(context.Background(), "", server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if feed.Title != "测试源" {
		t.Errorf("feed title = %q; want %q", feed.Title, "测试源")
	}
	if selected := model.Snapshot().CurrentFeedID; selected != feed.ID {
		t.Errorf("selected feed = %d; want %d", selected, feed.ID)
	}
}
