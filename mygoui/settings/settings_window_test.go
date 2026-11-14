package settings

import (
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

func TestSettingsViewRendersPreferences(t *testing.T) {
	prefs := app.DefaultPrefs()
	view := &settingsView{
		callbacks: Callbacks{
			Load:    func() app.Prefs { return prefs },
			Save:    func(next app.Prefs) (app.Prefs, error) { return next, nil },
			DataDir: func() string { return t.TempDir() },
			StorageUsage: func() app.StorageInfo {
				return app.StorageInfo{
					DBBytes:              48 << 20,
					ImageCacheBytes:      120 << 20,
					ImageCacheLimitBytes: 256 << 20,
				}
			},
			ExportOPML:  func(*mygo.Window) (string, error) { return "", nil },
			OpenDataDir: func() error { return nil },
		},
		prefs: prefs,
	}

	tester := ui.NewTester(view.draw, settingsWidth, settingsHeight)
	for _, want := range []string{
		"数据库大小：48.0 MB",
		"当前缓存占用：120.0 MB / 上限 256.0 MB",
	} {
		if !tester.HasText(want) {
			t.Errorf("Settings window is missing %q", want)
		}
	}
}
