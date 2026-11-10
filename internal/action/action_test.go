package action

import (
	"context"
	"errors"
	"testing"

	"github.com/sdttttt/ressu/internal/gui"
	"github.com/sdttttt/ressu/internal/service"
	"github.com/sdttttt/ressu/internal/store"
)

type fakeUI struct {
	openFiles    []string
	savePath     string
	openedPath   string
	externalLink string
	windowTitle  string
	windowLink   string
	theme        string
}

func (f *fakeUI) SelectOPMLImport(*gui.Window) ([]string, error) {
	return f.openFiles, nil
}

func (f *fakeUI) SelectOPMLExport(*gui.Window, string) (string, error) {
	return f.savePath, nil
}

func (f *fakeUI) SelectSnapshotSavePath(*gui.Window, string) (string, error) {
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

func (f *fakeUI) OpenImageViewer(*gui.Window, []byte, string, string) error {
	return nil
}

func (f *fakeUI) OpenSettings(*gui.Window, gui.SettingsCallbacks) {}

func (f *fakeUI) CapturePage(context.Context) ([]byte, error) {
	return []byte("png"), nil
}

func (f *fakeUI) WriteImage([]byte) error {
	return nil
}

func (f *fakeUI) ApplyTheme(theme string) {
	f.theme = theme
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
