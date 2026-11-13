package gui

import (
	"context"
	"testing"

	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/core/store"
	"github.com/sdttttt/ressu/internal/action"
)

// callbackFakeUI implements action.UI with no-op side effects.
type callbackFakeUI struct{}

func (callbackFakeUI) SelectOPMLImport(action.WindowHandle) ([]string, error) { return nil, nil }
func (callbackFakeUI) SelectOPMLExport(action.WindowHandle, string) (string, error) {
	return "", nil
}
func (callbackFakeUI) SelectSnapshotSavePath(action.WindowHandle, string) (string, error) {
	return "", nil
}
func (callbackFakeUI) OpenPath(string) error         { return nil }
func (callbackFakeUI) OpenExternal(string) error     { return nil }
func (callbackFakeUI) OpenLinkWindow(string, string) {}
func (callbackFakeUI) OpenImageViewer(action.WindowHandle, []byte, string, string) error {
	return nil
}
func (callbackFakeUI) OpenSnapshotViewer(action.WindowHandle, []byte, string, action.SnapshotViewerCallbacks) error {
	return nil
}
func (callbackFakeUI) OpenAddFeed(action.WindowHandle, action.AddFeedCallbacks) {}
func (callbackFakeUI) OpenEditFeed(action.WindowHandle, store.Feed, action.AddFeedCallbacks) {
}
func (callbackFakeUI) OpenSettings(action.WindowHandle, action.SettingsCallbacks) {}
func (callbackFakeUI) OpenSnapshotCapture(action.WindowHandle)                    {}
func (callbackFakeUI) CloseSnapshotCapture(action.WindowHandle)                   {}
func (callbackFakeUI) Confirm(context.Context, action.WindowHandle, action.ConfirmationRequest) (bool, error) {
	return false, nil
}
func (callbackFakeUI) CapturePage(context.Context) ([]byte, error) { return nil, nil }
func (callbackFakeUI) WriteImage([]byte) error                     { return nil }
func (callbackFakeUI) ApplyTheme(string)                           {}
func (callbackFakeUI) ShowNotification(action.Notification)        {}
func (callbackFakeUI) SelectNeighbourPost(int, bool)               {}
func (callbackFakeUI) SelectNeighbourFeed(int)                     {}
func (callbackFakeUI) ShowHelp()                                   {}

func TestSidebarCallbacksPersistLayout(t *testing.T) {
	st, err := store.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := st.Close(); err != nil {
			t.Error(err)
		}
	})
	actions := action.New(nil, nil, service.NewSettings(st), nil, nil, callbackFakeUI{})
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
