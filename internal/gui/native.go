package gui

import (
	"context"
	"fmt"

	"github.com/egoist/mygo"

	appaddfeed "github.com/sdttttt/ressu/internal/gui/addfeed"
	appdialog "github.com/sdttttt/ressu/internal/gui/dialog"
	apppostlist "github.com/sdttttt/ressu/internal/gui/postlist"
	appsettings "github.com/sdttttt/ressu/internal/gui/settings"
	appsidebar "github.com/sdttttt/ressu/internal/gui/sidebar"
	appsnapshotcap "github.com/sdttttt/ressu/internal/gui/snapshotcap"
	appviewer "github.com/sdttttt/ressu/internal/gui/viewer"
	"github.com/sdttttt/ressu/internal/store"
)

// NativeUI exposes application-level native side effects to the action layer.
type NativeUI struct {
	notifications *appdialog.Notifications
	confirmations *appdialog.Confirmations
	help          *appdialog.Help
	postList      *apppostlist.List
	sidebar       *appsidebar.Sidebar
}

// NewNativeUI creates the native side-effect adapter.
func NewNativeUI(
	notifications *appdialog.Notifications,
	confirmations *appdialog.Confirmations,
) *NativeUI {
	return &NativeUI{notifications: notifications, confirmations: confirmations}
}

// SetPostList connects the native article list to page-driven shortcuts.
func (n *NativeUI) SetPostList(postList *apppostlist.List) {
	n.postList = postList
}

// SetHelp connects the native keyboard shortcut modal to page actions.
func (n *NativeUI) SetHelp(help *appdialog.Help) {
	n.help = help
}

// SetSidebar connects the native subscription sidebar to page actions.
func (n *NativeUI) SetSidebar(sidebar *appsidebar.Sidebar) {
	n.sidebar = sidebar
}

// SelectNeighbourPost moves the native article list selection.
func (n *NativeUI) SelectNeighbourPost(delta int, onlyUnread bool) {
	if n.postList != nil {
		n.postList.SelectNeighbour(delta, onlyUnread)
	}
}

// SelectNeighbourFeed moves the native subscription selection.
func (n *NativeUI) SelectNeighbourFeed(delta int) {
	if n.help != nil && n.help.Visible() {
		return
	}
	if n.sidebar != nil {
		n.sidebar.SelectNeighbour(delta)
	}
}

// ShowHelp opens the native keyboard shortcut modal.
func (n *NativeUI) ShowHelp() {
	if n.help != nil {
		n.help.Show()
	}
}

// SelectOPMLImport shows the picker used while importing subscriptions.
func (NativeUI) SelectOPMLImport(parent *mygo.Window) ([]string, error) {
	return mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: parent,
		Title:  "选择 OPML 文件",
		Filters: []mygo.FileFilter{
			{Name: "OPML 订阅列表", Extensions: []string{"opml", "xml"}},
		},
	})
}

// SelectOPMLExport shows the save picker used while exporting subscriptions.
func (NativeUI) SelectOPMLExport(parent *mygo.Window, defaultPath string) (string, error) {
	return mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      parent,
		Title:       "导出 OPML 文件",
		DefaultPath: defaultPath,
		Filters: []mygo.FileFilter{
			{Name: "OPML 订阅列表", Extensions: []string{"opml"}},
		},
	})
}

// SelectSnapshotSavePath shows the save picker used by article snapshots.
func (NativeUI) SelectSnapshotSavePath(parent *mygo.Window, defaultPath string) (string, error) {
	return mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      parent,
		Title:       "保存文章快照",
		DefaultPath: defaultPath,
		Filters:     []mygo.FileFilter{{Name: "PNG 图片", Extensions: []string{"png"}}},
	})
}

// OpenPath shows a path in the file manager.
func (NativeUI) OpenPath(path string) error {
	return mygo.Shell.OpenPath(path)
}

// OpenExternal opens an HTTP or HTTPS link in the system browser.
func (NativeUI) OpenExternal(link string) error {
	return mygo.Shell.OpenExternal(link)
}

// OpenLinkWindow opens a link in an application window.
func (NativeUI) OpenLinkWindow(title, link string) {
	mygo.NewWindow(mygo.WindowOptions{Title: title, URL: link, Width: 1000, Height: 800})
}

// OpenImageViewer opens the native article image viewer.
func (NativeUI) OpenImageViewer(
	parent *mygo.Window,
	data []byte,
	viewerURL string,
	alt string,
) error {
	return appviewer.OpenImage(parent, data, viewerURL, alt)
}

// OpenSnapshotViewer opens the native generated-snapshot window.
func (NativeUI) OpenSnapshotViewer(
	parent *mygo.Window,
	data []byte,
	title string,
	callbacks appviewer.SnapshotCallbacks,
) error {
	return appviewer.OpenSnapshot(parent, data, title, callbacks)
}

// OpenSettings opens the native settings window.
func (NativeUI) OpenSettings(parent *mygo.Window, callbacks appsettings.Callbacks) {
	appsettings.Open(parent, callbacks)
}

// OpenSnapshotCapture opens the single snapshot capture window.
func (NativeUI) OpenSnapshotCapture(parent *mygo.Window) {
	appsnapshotcap.Open(parent)
}

// CloseSnapshotCapture closes the snapshot capture window.
func (NativeUI) CloseSnapshotCapture(window *mygo.Window) {
	appsnapshotcap.Close(window)
}

// OpenAddFeed opens the native add-feed window.
func (NativeUI) OpenAddFeed(parent *mygo.Window, callbacks appaddfeed.Callbacks) {
	appaddfeed.Open(parent, callbacks)
}

// Confirm shows an in-app confirmation for a destructive action.
func (n NativeUI) Confirm(
	ctx context.Context,
	parent *mygo.Window,
	request appdialog.ConfirmationRequest,
) (bool, error) {
	if n.confirmations == nil {
		return false, appdialog.ErrNoConfirmationSurface
	}
	return n.confirmations.Confirm(ctx, parent, request)
}

// CapturePage captures the visible caller page.
func (NativeUI) CapturePage(ctx context.Context) ([]byte, error) {
	if page := mygo.CallerPage(ctx); page != nil && page.WebView() != nil {
		return page.WebView().CapturePage()
	}
	window := mygo.CallerWindow(ctx)
	if window == nil || window.IsDestroyed() {
		return nil, fmt.Errorf("the snapshot window is no longer available")
	}
	return window.CapturePage()
}

// WriteImage writes an image to the system clipboard.
func (NativeUI) WriteImage(data []byte) error {
	return mygo.Clipboard.WriteImage(data)
}

// ApplyTheme applies the native appearance for a preference theme.
func (NativeUI) ApplyTheme(theme string) {
	switch theme {
	case store.ThemeLight:
		mygo.Theme.SetSource(mygo.ThemeLight)
	case store.ThemeDark, store.ThemeSepia:
		mygo.Theme.SetSource(mygo.ThemeDark)
	default:
		mygo.Theme.SetSource(mygo.ThemeSystem)
	}
}

// ShowNotification queues a native toast for the main window.
func (n NativeUI) ShowNotification(notification appdialog.Notification) {
	if n.notifications != nil {
		n.notifications.Show(notification)
	}
}
