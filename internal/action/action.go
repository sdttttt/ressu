// Package action defines the application operations shared by the web and
// native interfaces. Services keep the business logic; this layer owns UI
// side effects and state-change broadcasts.
package action

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/egoist/mygo"

	appevent "github.com/sdttttt/ressu/internal/event"
	"github.com/sdttttt/ressu/internal/gui"
	appaddfeed "github.com/sdttttt/ressu/internal/gui/addfeed"
	appdialog "github.com/sdttttt/ressu/internal/gui/dialog"
	apppostlist "github.com/sdttttt/ressu/internal/gui/postlist"
	appsettings "github.com/sdttttt/ressu/internal/gui/settings"
	appsidebar "github.com/sdttttt/ressu/internal/gui/sidebar"
	appviewer "github.com/sdttttt/ressu/internal/gui/viewer"
	"github.com/sdttttt/ressu/internal/service"
	"github.com/sdttttt/ressu/internal/store"
)

// ErrInvalidLink reports a link that is not an absolute HTTP or HTTPS URL.
var ErrInvalidLink = errors.New("invalid link")

// ErrNoSnapshotWindow reports that the snapshot caller window is unavailable.
var ErrNoSnapshotWindow = errors.New("no snapshot window")

// UI is the set of native side effects an action may perform.
type UI interface {
	SelectOPMLImport(parent *mygo.Window) ([]string, error)
	SelectOPMLExport(parent *mygo.Window, defaultPath string) (string, error)
	SelectSnapshotSavePath(parent *mygo.Window, defaultPath string) (string, error)
	OpenPath(string) error
	OpenExternal(string) error
	OpenLinkWindow(title, link string)
	OpenImageViewer(parent *mygo.Window, data []byte, viewerURL, alt string) error
	OpenSnapshotViewer(
		parent *mygo.Window,
		data []byte,
		title string,
		callbacks appviewer.SnapshotCallbacks,
	) error
	OpenAddFeed(parent *mygo.Window, callbacks appaddfeed.Callbacks)
	OpenSettings(parent *mygo.Window, callbacks appsettings.Callbacks)
	Confirm(ctx context.Context, parent *mygo.Window, request appdialog.ConfirmationRequest) (bool, error)
	CapturePage(context.Context) ([]byte, error)
	WriteImage([]byte) error
	ApplyTheme(string)
	ShowNotification(appdialog.Notification)
	SelectNeighbourPost(delta int, onlyUnread bool)
}

// AppActions is the single application-level entry for capabilities exposed to
// the WebView and native interfaces. Web bindings forward here instead of
// calling services directly.
type AppActions struct {
	feeds    *service.Feeds
	posts    *service.Posts
	settings *service.Settings
	status   *service.Status
	hub      *appevent.Hub
	ui       UI
}

// New creates the application action layer.
func New(
	feeds *service.Feeds,
	posts *service.Posts,
	settings *service.Settings,
	status *service.Status,
	hub *appevent.Hub,
	ui UI,
) *AppActions {
	return &AppActions{
		feeds:    feeds,
		posts:    posts,
		settings: settings,
		status:   status,
		hub:      hub,
		ui:       ui,
	}
}

// ListFeeds returns the subscriptions and total unread count.
func (a *AppActions) ListFeeds() service.FeedsSnapshot {
	return a.feeds.List()
}

// FeedIcon returns the cached icon of a feed.
func (a *AppActions) FeedIcon(ctx context.Context, id int64) (store.Icon, error) {
	return a.feeds.Icon(ctx, id)
}

// TestFeed validates and previews a subscription URL.
func (a *AppActions) TestFeed(ctx context.Context, feedURL string) (service.FeedPreview, error) {
	return a.feeds.Test(ctx, feedURL)
}

// AddFeed subscribes to a feed.
func (a *AppActions) AddFeed(ctx context.Context, title, feedURL string) (store.Feed, error) {
	return a.feeds.Add(ctx, title, feedURL)
}

// ConfirmRemoveFeed asks before deleting a subscription and its posts.
func (a *AppActions) ConfirmRemoveFeed(ctx context.Context, parent *mygo.Window, id int64) error {
	feed, err := a.feeds.Feed(id)
	if err != nil {
		return err
	}
	request := appdialog.ConfirmationRequest{
		Title:        "取消订阅「" + feed.Title + "」？",
		Description:  "这个订阅源已保存的文章会一并删除，操作无法撤销。",
		ConfirmLabel: "取消订阅",
		CancelLabel:  "取消",
		Destructive:  true,
	}
	confirmed, err := a.ui.Confirm(ctx, parent, request)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}
	return a.feeds.Remove(id)
}

// RefreshFeed fetches one subscription.
func (a *AppActions) RefreshFeed(ctx context.Context, id int64) (store.Feed, error) {
	return a.feeds.Refresh(ctx, id)
}

// RefreshAll fetches every subscription.
func (a *AppActions) RefreshAll(ctx context.Context) (service.SyncSummary, error) {
	return a.feeds.RefreshAll(ctx)
}

// RefreshStale fetches subscriptions that have not been fetched recently.
func (a *AppActions) RefreshStale(ctx context.Context) (service.SyncSummary, error) {
	return a.feeds.RefreshStale(ctx)
}

// ImportOPML asks for an OPML file and imports its subscriptions.
func (a *AppActions) ImportOPML(ctx context.Context, parent *mygo.Window) (service.ImportSummary, error) {
	paths, err := a.ui.SelectOPMLImport(parent)
	if err != nil {
		slog.Warn("OPML picker failed", "operation", "action.feeds.importOPML", "error", err)
		return service.ImportSummary{}, err
	}
	if len(paths) == 0 {
		return service.ImportSummary{}, nil
	}

	file, err := os.Open(paths[0])
	if err != nil {
		slog.Error("OPML open failed", "operation", "action.feeds.importOPML", "error", err)
		return service.ImportSummary{}, err
	}
	defer file.Close()
	return a.feeds.ImportOPML(ctx, file)
}

// ExportOPML asks where to save the subscriptions and writes an OPML file.
func (a *AppActions) ExportOPML(parent *mygo.Window) (string, error) {
	data, err := a.feeds.BuildOPML()
	if err != nil {
		return "", err
	}
	path, err := a.ui.SelectOPMLExport(parent, "ressu-subscriptions.opml")
	if err != nil {
		slog.Warn("OPML save picker failed", "operation", "action.feeds.exportOPML", "error", err)
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		slog.Error("OPML write failed", "operation", "action.feeds.exportOPML", "error", err)
		return "", fmt.Errorf("could not save the OPML file: %w", err)
	}
	return path, nil
}

// OpenImage loads an article image and opens it in the native viewer.
func (a *AppActions) OpenImage(ctx context.Context, parent *mygo.Window, source, alt string) error {
	image, err := a.posts.LoadImage(ctx, source)
	if err != nil {
		return err
	}
	return a.ui.OpenImageViewer(parent, image.Data, image.ViewerURL, alt)
}

// ListPosts returns a page of posts.
func (a *AppActions) ListPosts(query service.PostQuery) (service.PostPage, error) {
	return a.posts.List(query)
}

// GetPost returns one complete post.
func (a *AppActions) GetPost(feedID int64, guid string) (store.Post, error) {
	return a.posts.Get(feedID, guid)
}

// MarkPostRead changes one post's read state.
func (a *AppActions) MarkPostRead(feedID int64, guid string, read bool) error {
	return a.posts.MarkRead(feedID, guid, read)
}

// MarkAllPostsRead marks every post of a feed as read.
func (a *AppActions) MarkAllPostsRead(feedID int64) (int, error) {
	return a.posts.MarkAllRead(feedID)
}

// TogglePostStar changes one post's starred state.
func (a *AppActions) TogglePostStar(feedID int64, guid string) (store.PostSummary, error) {
	return a.posts.ToggleStar(feedID, guid)
}

// CaptureSnapshot captures the visible page for one scroll position.
func (a *AppActions) CaptureSnapshot(ctx context.Context, request service.SnapshotCaptureRequest) ([]byte, error) {
	if err := service.ValidateSnapshotCapture(request); err != nil {
		return nil, err
	}
	data, err := a.ui.CapturePage(ctx)
	if err != nil {
		slog.Warn("snapshot capture failed",
			"operation", "action.posts.captureSnapshot", "error", err)
	}
	return data, err
}

// StitchSnapshot joins page captures and opens the native snapshot preview.
func (a *AppActions) StitchSnapshot(
	ctx context.Context,
	title string,
	tiles []service.SnapshotTile,
) error {
	parent := mygo.CallerWindow(ctx)
	if parent == nil {
		return ErrNoSnapshotWindow
	}
	return a.stitchSnapshot(parent, title, tiles)
}

// stitchSnapshot joins page captures and opens the native preview for parent.
func (a *AppActions) stitchSnapshot(
	parent *mygo.Window,
	title string,
	tiles []service.SnapshotTile,
) error {
	data, err := a.posts.StitchSnapshot(context.Background(), tiles)
	if err != nil {
		return err
	}
	return a.ui.OpenSnapshotViewer(parent, data, title, appviewer.SnapshotCallbacks{
		Save: func(snapshotParent *mygo.Window, snapshotData []byte, fileName string) (string, error) {
			return a.SaveSnapshot(snapshotParent, fileName, snapshotData)
		},
		Copy:   a.CopySnapshot,
		Notify: a.Notify,
	})
}

// SaveSnapshot asks where to save a stitched snapshot and writes it.
func (a *AppActions) SaveSnapshot(parent *mygo.Window, fileName string, data []byte) (string, error) {
	if err := service.ValidateSnapshotImage(data); err != nil {
		return "", err
	}
	path, err := a.ui.SelectSnapshotSavePath(parent, service.SnapshotFileName(fileName))
	if err != nil {
		slog.Warn("snapshot save picker failed", "operation", "action.posts.saveSnapshot", "error", err)
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.Error("snapshot write failed", "operation", "action.posts.saveSnapshot", "error", err)
		return "", fmt.Errorf("could not save the snapshot: %w", err)
	}
	return path, nil
}

// CopySnapshot puts a stitched snapshot on the clipboard.
func (a *AppActions) CopySnapshot(data []byte) error {
	if err := service.ValidateSnapshotImage(data); err != nil {
		slog.Warn("invalid snapshot image", "operation", "action.posts.copySnapshot", "error", err)
		return err
	}
	if err := a.ui.WriteImage(data); err != nil {
		slog.Warn("snapshot copy failed", "operation", "action.posts.copySnapshot", "error", err)
		return err
	}
	return nil
}

// Preferences returns the current preferences.
func (a *AppActions) Preferences() store.Prefs {
	return a.settings.Get()
}

// UpdatePreferences stores preferences and broadcasts every affected change.
func (a *AppActions) UpdatePreferences(prefs store.Prefs) (store.Prefs, error) {
	stored, err := a.settings.Update(prefs)
	if err != nil {
		return store.Prefs{}, err
	}
	a.ui.ApplyTheme(stored.Theme)
	if a.hub != nil {
		a.hub.AppearanceChanged(stored.Theme, stored.FontSize)
		a.hub.PrefsChanged(stored)
	}
	return stored, nil
}

// UpdateSettingsDraft saves a native settings draft while preserving layout
// fields that the native surface does not edit.
func (a *AppActions) UpdateSettingsDraft(prefs store.Prefs) (store.Prefs, error) {
	current := a.settings.Get()
	prefs.PaneSizes = append([]int(nil), current.PaneSizes...)
	prefs.SidebarCollapsed = current.SidebarCollapsed
	return a.UpdatePreferences(prefs)
}

// DataDir returns the app's data directory.
func (a *AppActions) DataDir() string {
	return a.settings.DataDir()
}

// OpenDataDir shows the data directory in the file manager.
func (a *AppActions) OpenDataDir() error {
	if err := a.ui.OpenPath(a.settings.DataDir()); err != nil {
		slog.Warn("data directory open failed",
			"operation", "action.settings.openDataDir", "error", err)
		return err
	}
	return nil
}

// OpenLink opens an absolute HTTP or HTTPS link according to the preferences.
func (a *AppActions) OpenLink(link string) error {
	link = strings.TrimSpace(link)
	if link == "" {
		return nil
	}
	parsed, err := url.Parse(link)
	if err != nil || parsed.Host == "" ||
		(strings.ToLower(parsed.Scheme) != "http" && strings.ToLower(parsed.Scheme) != "https") {
		return ErrInvalidLink
	}
	if a.settings.Get().OpenLinks == store.OpenLinksWindow {
		a.ui.OpenLinkWindow(link, link)
		return nil
	}
	if err := a.ui.OpenExternal(link); err != nil {
		slog.Warn("external link open failed",
			"operation", "action.settings.openLink", "externalHost", parsed.Host, "error", err)
		return err
	}
	return nil
}

// OpenSettings opens the native settings surface with action callbacks.
func (a *AppActions) OpenSettings(parent *mygo.Window) {
	a.ui.OpenSettings(parent, appsettings.Callbacks{
		Load:        a.settings.Get,
		Save:        a.UpdateSettingsDraft,
		DataDir:     a.settings.DataDir,
		ExportOPML:  a.ExportOPML,
		OpenDataDir: a.OpenDataDir,
	})
}

// OpenAddFeed opens the native subscription surface.
func (a *AppActions) OpenAddFeed(parent *mygo.Window) {
	a.ui.OpenAddFeed(parent, appaddfeed.Callbacks{
		Appearance: a.settings.Get,
		Test: func(ctx context.Context, feedURL string) (appaddfeed.Preview, error) {
			preview, err := a.TestFeed(ctx, feedURL)
			return appaddfeed.Preview(preview), err
		},
		Add: a.AddAndSelectFeed,
		ImportOPML: func(feedParent *mygo.Window) (appaddfeed.Summary, error) {
			summary, err := a.ImportOPML(context.Background(), feedParent)
			return appaddfeed.Summary(summary), err
		},
	})
}

// SidebarCallbacks exposes native subscription actions and layout updates.
func (a *AppActions) SidebarCallbacks() appsidebar.Callbacks {
	return appsidebar.Callbacks{
		Icon:        a.FeedIcon,
		Select:      func(feedID int64) { a.SelectFeed(context.Background(), feedID) },
		RefreshAll:  func() { a.refreshAllWithNotice() },
		RefreshFeed: func(feedID int64) { a.refreshFeedWithNotice(feedID) },
		MarkAllRead: func(feedID int64) { a.markAllReadWithNotice(feedID) },
		Remove: func(parent *mygo.Window, feedID int64) {
			if err := a.ConfirmRemoveFeed(context.Background(), parent, feedID); err != nil {
				slog.Warn("native feed removal failed",
					"operation", "action.sidebar.remove", "feedID", feedID, "error", err)
				a.Notify("error", "取消订阅失败")
			}
		},
		OpenAddFeed:    a.OpenAddFeed,
		OpenSettings:   a.OpenSettings,
		SetCollapsed:   a.SetSidebarCollapsed,
		SetPanePercent: a.SetSidebarPercent,
	}
}

// PostListCallbacks exposes native article list actions and layout updates.
func (a *AppActions) PostListCallbacks() apppostlist.Callbacks {
	return apppostlist.Callbacks{
		List:           a.ListPosts,
		Select:         a.SelectPost,
		ToggleStar:     a.TogglePostStar,
		MarkAllRead:    func(feedID int64) { a.markAllReadWithNotice(feedID) },
		SetPanePercent: a.SetListPercent,
		Notify:         a.Notify,
	}
}

// PostHeaderCallbacks exposes native article header actions.
func (a *AppActions) PostHeaderCallbacks() gui.PostHeaderCallbacks {
	return gui.PostHeaderCallbacks{
		ToggleStar:      a.TogglePostStar,
		MarkRead:        a.MarkPostRead,
		OpenLink:        a.OpenLink,
		RequestSnapshot: a.RequestSnapshot,
		Notify:          a.Notify,
	}
}

// RequestSnapshot asks the reading page to generate the selected post snapshot.
func (a *AppActions) RequestSnapshot() {
	a.hub.UICommand("snapshot")
}

func (a *AppActions) refreshAllWithNotice() {
	if _, err := a.RefreshAll(context.Background()); err != nil {
		if errors.Is(err, service.ErrRefreshing) {
			return
		}
		slog.Warn("native refresh all failed",
			"operation", "action.sidebar.refreshAll", "error", err)
		a.Notify("error", "刷新订阅源失败")
	}
}

func (a *AppActions) refreshFeedWithNotice(feedID int64) {
	if _, err := a.RefreshFeed(context.Background(), feedID); err != nil {
		slog.Warn("native feed refresh failed",
			"operation", "action.sidebar.refreshFeed", "feedID", feedID, "error", err)
		a.Notify("error", "刷新订阅源失败")
	}
}

func (a *AppActions) markAllReadWithNotice(feedID int64) {
	if _, err := a.MarkAllPostsRead(feedID); err != nil {
		slog.Warn("native mark all read failed",
			"operation", "action.sidebar.markAllRead", "feedID", feedID, "error", err)
		a.Notify("error", "标记已读失败")
	}
}

// SetSidebarCollapsed stores the native sidebar's collapsed preference.
func (a *AppActions) SetSidebarCollapsed(collapsed bool) {
	prefs := a.settings.Get()
	if prefs.SidebarCollapsed == collapsed {
		return
	}
	prefs.SidebarCollapsed = collapsed
	if _, err := a.UpdatePreferences(prefs); err != nil {
		slog.Warn("native sidebar collapse save failed",
			"operation", "action.sidebar.setCollapsed", "collapsed", collapsed, "error", err)
		a.Notify("error", "保存侧边栏状态失败")
	}
}

// SetSidebarPercent stores the native sidebar's docked width percentage.
func (a *AppActions) SetSidebarPercent(percent float32) {
	prefs := a.settings.Get()
	sideWidth := min(max(int(math.Round(float64(percent))), 5), 90)
	listWidth := 30
	if len(prefs.PaneSizes) == 2 {
		listWidth = prefs.PaneSizes[1]
	}
	if len(prefs.PaneSizes) == 2 && prefs.PaneSizes[0] == sideWidth {
		return
	}
	prefs.PaneSizes = []int{sideWidth, listWidth}
	if _, err := a.UpdatePreferences(prefs); err != nil {
		slog.Warn("native sidebar width save failed",
			"operation", "action.sidebar.setPanePercent", "percent", percent, "error", err)
		a.Notify("error", "保存侧边栏宽度失败")
	}
}

// SetListPercent stores the native article list's docked width percentage.
func (a *AppActions) SetListPercent(percent float32) {
	prefs := a.settings.Get()
	listWidth := min(max(int(math.Round(float64(percent))), 18), 55)
	sideWidth := 18
	if len(prefs.PaneSizes) == 2 {
		sideWidth = prefs.PaneSizes[0]
	}
	if len(prefs.PaneSizes) == 2 && prefs.PaneSizes[1] == listWidth {
		return
	}
	prefs.PaneSizes = []int{sideWidth, listWidth}
	if _, err := a.UpdatePreferences(prefs); err != nil {
		slog.Warn("native list width save failed",
			"operation", "action.postList.setPanePercent", "percent", percent, "error", err)
		a.Notify("error", "保存文章栏宽度失败")
	}
}

// SelectFeed broadcasts the feed selected by a page.
func (a *AppActions) SelectFeed(ctx context.Context, feedID int64) {
	a.status.SelectFeed(ctx, feedID)
}

// SelectPost broadcasts the article selected by the native list.
func (a *AppActions) SelectPost(post store.PostSummary) {
	a.status.SelectPost(context.Background(), post)
}

// AddAndSelectFeed subscribes to a feed and broadcasts its selection.
func (a *AppActions) AddAndSelectFeed(ctx context.Context, title, feedURL string) (store.Feed, error) {
	feed, err := a.AddFeed(ctx, title, feedURL)
	if err != nil {
		return store.Feed{}, err
	}
	a.SelectFeed(ctx, feed.ID)
	return feed, nil
}

// PageReady marks the reader page as ready to receive a full snapshot.
func (a *AppActions) PageReady(ctx context.Context) {
	a.status.PageReady(ctx)
}

// Notify shows a native application notification.
func (a *AppActions) Notify(level, message string) {
	a.ui.ShowNotification(appdialog.Notification{Type: level, Title: message})
}

// SelectNeighbourPost moves the native article list from a page shortcut.
func (a *AppActions) SelectNeighbourPost(delta int, onlyUnread bool) {
	a.ui.SelectNeighbourPost(delta, onlyUnread)
}
