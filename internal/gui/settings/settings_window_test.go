package settings

import (
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/store"
)

func TestSettingsViewRendersPreferences(t *testing.T) {
	prefs := store.DefaultPrefs()
	view := &settingsView{
		callbacks: Callbacks{
			Load:        func() store.Prefs { return prefs },
			Save:        func(next store.Prefs) (store.Prefs, error) { return next, nil },
			DataDir:     func() string { return t.TempDir() },
			ExportOPML:  func(*mygo.Window) (string, error) { return "", nil },
			OpenDataDir: func() error { return nil },
		},
		prefs: prefs,
	}

	ui.NewTester(view.draw, settingsWidth, settingsHeight)
}
