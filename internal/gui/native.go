package gui

import (
	"context"
	"fmt"

	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/store"
)

// Window is the native window type used by GUI entry points.
type Window = mygo.Window

// SettingsCallbacks are the operations a native settings surface may invoke.
type SettingsCallbacks struct {
	Load        func() store.Prefs
	Save        func(store.Prefs) (store.Prefs, error)
	DataDir     func() string
	ExportOPML  func(parent *Window) (string, error)
	OpenDataDir func() error
}

// NativeUI exposes application-level native side effects to the action layer.
type NativeUI struct {
	notifications *Notifications
}

// NewNativeUI creates the native side-effect adapter.
func NewNativeUI(notifications *Notifications) NativeUI {
	return NativeUI{notifications: notifications}
}

// SelectOPMLImport shows the picker used while importing subscriptions.
func (NativeUI) SelectOPMLImport(parent *Window) ([]string, error) {
	return mygo.Dialog.Open(mygo.OpenDialogOptions{
		Parent: parent,
		Title:  "选择 OPML 文件",
		Filters: []mygo.FileFilter{
			{Name: "OPML 订阅列表", Extensions: []string{"opml", "xml"}},
		},
	})
}

// SelectOPMLExport shows the save picker used while exporting subscriptions.
func (NativeUI) SelectOPMLExport(parent *Window, defaultPath string) (string, error) {
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
func (NativeUI) SelectSnapshotSavePath(parent *Window, defaultPath string) (string, error) {
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
	parent *Window,
	data []byte,
	viewerURL string,
	alt string,
) error {
	return OpenImageViewer(parent, data, viewerURL, alt)
}

// OpenSettings opens the native settings window.
func (NativeUI) OpenSettings(parent *Window, callbacks SettingsCallbacks) {
	OpenSettingsWindow(parent, callbacks)
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
func (n NativeUI) ShowNotification(notification Notification) {
	if n.notifications != nil {
		n.notifications.Show(notification)
	}
}
