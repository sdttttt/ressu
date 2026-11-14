// Package webendpoint exposes application actions to the reader page through
// MyGO bindings. It only adapts caller contexts and delegates to actions.
package webendpoint

import (
	"context"
	"errors"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/core/app"
)

// AppEvent is the only event channel the Go side broadcasts to pages. The Hub
// sinks into it so the event package stays free of UI-framework imports.
var AppEvent = mygo.NewEvent[app.Event]("app:event")

// ErrNoSnapshotWindow reports that the snapshot caller window is unavailable.
var ErrNoSnapshotWindow = errors.New("no snapshot window")

// WebEndpoint is the page binding for subscription actions. It is exposed to
// the page as Feeds.
type WebEndpoint struct {
	actions *app.App
}

// NewWebEndpoint creates the subscription endpoint.
func NewWebEndpoint(actions *app.App) *WebEndpoint {
	return &WebEndpoint{actions: actions}
}

// List returns the subscriptions and total unread count.
func (w *WebEndpoint) List() app.FeedsSnapshot {
	return w.actions.ListFeeds()
}

// Icon returns a feed's cached icon.
func (w *WebEndpoint) Icon(ctx context.Context, id int64) (app.Icon, error) {
	return w.actions.FeedIcon(ctx, id)
}

// Test validates and previews a subscription URL.
func (w *WebEndpoint) Test(ctx context.Context, feedURL string) (app.FeedPreview, error) {
	return w.actions.TestFeed(ctx, feedURL, 0)
}

// Add subscribes to a feed with the global default remote-resource mode.
func (w *WebEndpoint) Add(ctx context.Context, title, feedURL string) (app.Feed, error) {
	return w.actions.AddFeed(ctx, title, feedURL, "")
}

// OpenAddFeed opens the native subscription surface.
func (w *WebEndpoint) OpenAddFeed(ctx context.Context) {
	w.actions.OpenAddFeed(app.NewWindowHandle(mygo.CallerWindow(ctx)))
}

// ConfirmRemove asks before deleting a subscription and its posts.
func (w *WebEndpoint) ConfirmRemove(ctx context.Context, id int64) error {
	return w.actions.ConfirmRemoveFeed(ctx, app.NewWindowHandle(mygo.CallerWindow(ctx)), id)
}

// Refresh fetches one subscription.
func (w *WebEndpoint) Refresh(ctx context.Context, id int64) (app.Feed, error) {
	return w.actions.RefreshFeed(ctx, id)
}

// RefreshAll fetches every subscription.
func (w *WebEndpoint) RefreshAll(ctx context.Context) (app.SyncSummary, error) {
	return w.actions.RefreshAll(ctx)
}

// RefreshStale fetches subscriptions that have not been fetched recently.
func (w *WebEndpoint) RefreshStale(ctx context.Context) (app.SyncSummary, error) {
	return w.actions.RefreshStale(ctx)
}

// ImportOPML asks for an OPML file and imports its subscriptions.
func (w *WebEndpoint) ImportOPML(ctx context.Context) (app.ImportSummary, error) {
	return w.actions.ImportOPML(ctx, app.NewWindowHandle(mygo.CallerWindow(ctx)))
}

// ExportOPML asks where to save the subscriptions and writes an OPML file.
func (w *WebEndpoint) ExportOPML(ctx context.Context) (string, error) {
	return w.actions.ExportOPML(app.NewWindowHandle(mygo.CallerWindow(ctx)))
}

// Posts is the page binding for article actions.
type Posts struct {
	actions *app.App
}

// NewPosts creates the article endpoint.
func NewPosts(actions *app.App) *Posts {
	return &Posts{actions: actions}
}

// OpenImage loads an article image into the native viewer.
func (p *Posts) OpenImage(ctx context.Context, source, alt string) error {
	return p.actions.OpenImage(ctx, app.NewWindowHandle(mygo.CallerWindow(ctx)), source, alt)
}

// List returns a page of posts.
func (p *Posts) List(query app.PostQuery) (app.PostPage, error) {
	return p.actions.ListPosts(query)
}

// Get returns one complete post.
func (p *Posts) Get(feedID int64, guid string) (app.Post, error) {
	return p.actions.GetPost(feedID, guid)
}

// MarkAllRead marks every post of a feed as read.
func (p *Posts) MarkAllRead(feedID int64) (int, error) {
	return p.actions.MarkAllPostsRead(feedID)
}

// ToggleRead flips one post's read state.
func (p *Posts) ToggleRead(feedID int64, guid string) (app.PostSummary, error) {
	return p.actions.TogglePostRead(feedID, guid)
}

// ToggleStar changes one post's starred state.
func (p *Posts) ToggleStar(feedID int64, guid string) (app.PostSummary, error) {
	return p.actions.TogglePostStar(feedID, guid)
}

// SelectNeighbour moves the native article list selection.
func (p *Posts) SelectNeighbour(_ context.Context, delta int, onlyUnread bool) {
	p.actions.SelectNeighbourPost(delta, onlyUnread)
}

// CaptureSnapshot captures the visible page for one scroll position.
func (p *Posts) CaptureSnapshot(ctx context.Context, request app.SnapshotCaptureRequest) ([]byte, error) {
	return p.actions.CaptureSnapshot(ctx, request)
}

// StitchSnapshot joins page captures and opens the native preview.
// The caller is the capture window; the preview parents to the main window
// and the capture window closes once the preview is open.
func (p *Posts) StitchSnapshot(
	ctx context.Context,
	title string,
	tiles []app.SnapshotTile,
) error {
	caller := mygo.CallerWindow(ctx)
	if caller == nil {
		return ErrNoSnapshotWindow
	}
	parent := caller
	if owner := caller.Parent(); owner != nil && !owner.IsDestroyed() {
		parent = owner
	}
	err := p.actions.StitchSnapshot(app.NewWindowHandle(parent), title, tiles)
	if err == nil && parent != caller && !caller.IsDestroyed() {
		caller.Close()
	}
	return err
}

// SnapshotTarget identifies the article the capture window renders.
type SnapshotTarget struct {
	FeedID    int64  `json:"feedId"`
	GUID      string `json:"guid"`
	FeedTitle string `json:"feedTitle"`
}

// Snapshot is the page binding of the snapshot capture window.
type Snapshot struct {
	actions *app.App
}

// NewSnapshot creates the snapshot endpoint.
func NewSnapshot(actions *app.App) *Snapshot {
	return &Snapshot{actions: actions}
}

// Selected returns the article the capture window should render.
func (s *Snapshot) Selected() SnapshotTarget {
	selection, feedTitle, ok := s.actions.SnapshotTarget()
	if !ok {
		return SnapshotTarget{}
	}
	return SnapshotTarget{
		FeedID:    selection.FeedID,
		GUID:      selection.GUID,
		FeedTitle: feedTitle,
	}
}

// Cancel closes the snapshot capture window.
func (s *Snapshot) Cancel(ctx context.Context) {
	s.actions.CancelSnapshotCapture(app.NewWindowHandle(mygo.CallerWindow(ctx)))
}

// Settings is the page binding for preferences and shell actions.
type Settings struct {
	actions *app.App
}

// NewSettings creates the settings endpoint.
func NewSettings(actions *app.App) *Settings {
	return &Settings{actions: actions}
}

// Get returns the current preferences.
func (s *Settings) Get() app.Prefs {
	return s.actions.Preferences()
}

// Update stores preferences.
func (s *Settings) Update(prefs app.Prefs) (app.Prefs, error) {
	return s.actions.UpdatePreferences(prefs)
}

// DataDir returns the app's data directory.
func (s *Settings) DataDir() string {
	return s.actions.DataDir()
}

// OpenDataDir shows the data directory in the file manager.
func (s *Settings) OpenDataDir() error {
	return s.actions.OpenDataDir()
}

// OpenLink opens an external link.
func (s *Settings) OpenLink(ctx context.Context, link string) error {
	return s.actions.OpenLink(link)
}

// OpenSettings opens the native settings window.
func (s *Settings) OpenSettings(ctx context.Context) {
	s.actions.OpenSettings(app.NewWindowHandle(mygo.CallerWindow(ctx)))
}

// Status is the page binding for native status state.
type Status struct {
	actions *app.App
}

// NewStatus creates the status endpoint.
func NewStatus(actions *app.App) *Status {
	return &Status{actions: actions}
}

// SelectFeed broadcasts the selected feed.
func (s *Status) SelectFeed(ctx context.Context, feedID int64) {
	s.actions.SelectFeed(ctx, feedID)
}

// SelectNeighbourFeed moves the native subscription selection.
func (s *Status) SelectNeighbourFeed(_ context.Context, delta int) {
	s.actions.SelectNeighbourFeed(delta)
}

// ShowHelp opens the native keyboard shortcut modal.
func (s *Status) ShowHelp(_ context.Context) {
	s.actions.ShowHelp()
}

// PageReady marks the reader page as ready for a full snapshot.
func (s *Status) PageReady(ctx context.Context) {
	s.actions.PageReady(ctx)
}

// Notify shows a native application notification.
func (s *Status) Notify(level, message string) {
	s.actions.Notify(level, message)
}
