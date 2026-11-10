// Package action defines the application operations shared by the web and
// native interfaces. Services keep the business logic; this layer owns UI
// side effects and state-change broadcasts.
package action

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	appevent "github.com/sdttttt/ressu/internal/event"
	"github.com/sdttttt/ressu/internal/gui"
	"github.com/sdttttt/ressu/internal/service"
	"github.com/sdttttt/ressu/internal/store"
)

// ErrInvalidLink reports a link that is not an absolute HTTP or HTTPS URL.
var ErrInvalidLink = errors.New("invalid link")

// UI is the set of native side effects an action may perform.
type UI interface {
	SelectOPMLImport(parent *gui.Window) ([]string, error)
	SelectOPMLExport(parent *gui.Window, defaultPath string) (string, error)
	SelectSnapshotSavePath(parent *gui.Window, defaultPath string) (string, error)
	OpenPath(string) error
	OpenExternal(string) error
	OpenLinkWindow(title, link string)
	OpenImageViewer(parent *gui.Window, data []byte, viewerURL, alt string) error
	OpenSettings(parent *gui.Window, callbacks gui.SettingsCallbacks)
	CapturePage(context.Context) ([]byte, error)
	WriteImage([]byte) error
	ApplyTheme(string)
	ShowNotification(gui.Notification)
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

// RemoveFeed deletes a subscription.
func (a *AppActions) RemoveFeed(id int64) error {
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
func (a *AppActions) ImportOPML(ctx context.Context, parent *gui.Window) (service.ImportSummary, error) {
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
func (a *AppActions) ExportOPML(parent *gui.Window) (string, error) {
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
func (a *AppActions) OpenImage(ctx context.Context, parent *gui.Window, source, alt string) error {
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

// StitchSnapshot joins page captures into one PNG.
func (a *AppActions) StitchSnapshot(ctx context.Context, tiles []service.SnapshotTile) ([]byte, error) {
	return a.posts.StitchSnapshot(ctx, tiles)
}

// SaveSnapshot asks where to save a stitched snapshot and writes it.
func (a *AppActions) SaveSnapshot(parent *gui.Window, fileName string, data []byte) (string, error) {
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
func (a *AppActions) OpenSettings(parent *gui.Window) {
	a.ui.OpenSettings(parent, gui.SettingsCallbacks{
		Load:        a.settings.Get,
		Save:        a.UpdateSettingsDraft,
		DataDir:     a.settings.DataDir,
		ExportOPML:  a.ExportOPML,
		OpenDataDir: a.OpenDataDir,
	})
}

// SelectFeed broadcasts the feed selected by a page.
func (a *AppActions) SelectFeed(ctx context.Context, feedID int64) {
	a.status.SelectFeed(ctx, feedID)
}

// PageReady marks the reader page as ready to receive a full snapshot.
func (a *AppActions) PageReady(ctx context.Context) {
	a.status.PageReady(ctx)
}

// Notify shows a native application notification.
func (a *AppActions) Notify(level, message string) {
	a.ui.ShowNotification(gui.Notification{Type: level, Title: message})
}
