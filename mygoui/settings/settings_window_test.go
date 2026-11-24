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

func TestSettingsViewRendersFontAndReadingFields(t *testing.T) {
	prefs := app.DefaultPrefs()
	view := &settingsView{
		callbacks: Callbacks{
			Load:    func() app.Prefs { return prefs },
			Save:    func(next app.Prefs) (app.Prefs, error) { return next, nil },
			DataDir: func() string { return t.TempDir() },
		},
		prefs:            prefs,
		uiFontDraft:      fontFamilyLabel(prefs.UIFontFamily),
		articleFontDraft: fontFamilyLabel(prefs.ArticleFontFamily),
	}

	tester := ui.NewTester(view.draw, settingsWidth, settingsHeight)
	for _, want := range []string{"界面字体", "文章字体", "文章字号", "文章行高"} {
		if !tester.HasText(want) {
			t.Errorf("Settings window is missing %q", want)
		}
	}
}

func TestFontFamilyLabelAndValue(t *testing.T) {
	if got := fontFamilyLabel(app.FontFamilySystem); got != "系统默认" {
		t.Errorf("fontFamilyLabel(system) = %q", got)
	}
	if got := fontFamilyLabel(app.FontFamilySerif); got != "衬线" {
		t.Errorf("fontFamilyLabel(serif) = %q", got)
	}
	custom := `"Microsoft YaHei", Arial`
	if got := fontFamilyLabel(custom); got != custom {
		t.Errorf("fontFamilyLabel(custom) = %q, want %q", got, custom)
	}

	if got := fontFamilyValue("等宽"); got != app.FontFamilyMono {
		t.Errorf("fontFamilyValue(等宽) = %q", got)
	}
	if got := fontFamilyValue("  衬线  "); got != app.FontFamilySerif {
		t.Errorf("fontFamilyValue(trimmed label) = %q", got)
	}
	if got := fontFamilyValue(""); got != app.FontFamilySystem {
		t.Errorf("fontFamilyValue(empty) = %q, want %q", got, app.FontFamilySystem)
	}
	if got := fontFamilyValue(custom); got != custom {
		t.Errorf("fontFamilyValue(custom) = %q, want %q", got, custom)
	}
}
