package mygoui

import (
	"context"
	"testing"

	"github.com/sdttttt/ressu/core/app"
)

// callbackFakeUI implements app.UI with no-op side effects.
type callbackFakeUI struct{}

func (callbackFakeUI) SelectOPMLImport(app.WindowHandle) ([]string, error) { return nil, nil }
func (callbackFakeUI) SelectOPMLExport(app.WindowHandle, string) (string, error) {
	return "", nil
}
func (callbackFakeUI) SelectSnapshotSavePath(app.WindowHandle, string) (string, error) {
	return "", nil
}
func (callbackFakeUI) OpenPath(string) error         { return nil }
func (callbackFakeUI) OpenExternal(string) error     { return nil }
func (callbackFakeUI) OpenLinkWindow(string, string) {}
func (callbackFakeUI) OpenImageViewer(app.WindowHandle, []byte, string, string) error {
	return nil
}
func (callbackFakeUI) OpenSnapshotViewer(app.WindowHandle, []byte, string, app.SnapshotViewerCallbacks) error {
	return nil
}
func (callbackFakeUI) OpenAddFeed(app.WindowHandle, app.AddFeedCallbacks) {}
func (callbackFakeUI) OpenEditFeed(app.WindowHandle, app.Feed, app.AddFeedCallbacks) {
}
func (callbackFakeUI) OpenSettings(app.WindowHandle, app.SettingsCallbacks) {}
func (callbackFakeUI) OpenSnapshotCapture(app.WindowHandle)                 {}
func (callbackFakeUI) CloseSnapshotCapture(app.WindowHandle)                {}
func (callbackFakeUI) Confirm(context.Context, app.WindowHandle, app.ConfirmationRequest) (bool, error) {
	return false, nil
}
func (callbackFakeUI) CapturePage(context.Context) ([]byte, error) { return nil, nil }
func (callbackFakeUI) WriteImage([]byte) error                     { return nil }
func (callbackFakeUI) ApplyTheme(string)                           {}
func (callbackFakeUI) ShowNotification(app.Notification)           {}
func (callbackFakeUI) SelectNeighbourPost(int, bool)               {}
func (callbackFakeUI) SelectNeighbourFeed(int)                     {}
func (callbackFakeUI) ShowHelp()                                   {}

func TestSidebarCallbacksPersistLayout(t *testing.T) {
	st, err := app.OpenStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})
	actions := app.New(nil, nil, app.NewSettings(st), nil, nil, callbackFakeUI{})
	callbacks := SidebarCallbacks(actions)

	callbacks.SetCollapsed(true)
	callbacks.SetPanePercent(27)

	prefs := actions.Preferences()
	if !prefs.SidebarCollapsed {
		t.Error("collapsed preference was not stored")
	}
	if len(prefs.PaneSizes) != 2 || prefs.PaneSizes[0] != 27 || prefs.PaneSizes[1] != 30 {
		t.Errorf("pane sizes = %v; want [27 30]", prefs.PaneSizes)
	}
}
