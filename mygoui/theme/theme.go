// Package theme maps application appearance preferences to native UI themes.
package theme

import (
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

// StatusBarHeight returns the rendered status bar height for a font size.
func StatusBarHeight(fontSize string) float32 {
	return 1 + StatusBarContentHeight*FontScale(fontSize)
}

// Appearance bundles the preference values that shape a native theme.
type Appearance struct {
	Theme      string
	FontSize   string
	FontFamily string // UIFontFamily preference: preset keyword or custom CSS list
}

// New maps appearance preferences to the colors and fonts used by native UI.
func New(c *ui.Context, appearance Appearance) *ui.Theme {
	base := ui.LightTheme()
	switch appearance.Theme {
	case app.ThemeDark:
		base = ui.DarkTheme()
	case app.ThemeSepia:
		base = ui.LightTheme()
	case app.ThemeSystem:
		if c.Theme().Dark {
			base = ui.DarkTheme()
		}
	}

	nativeTheme := *base
	nativeTheme.FontSize *= FontScale(appearance.FontSize)
	nativeTheme.Font = app.ResolveFontFamily(appearance.FontFamily)
	if appearance.Theme == app.ThemeSepia {
		ApplySepia(&nativeTheme)
	}
	return &nativeTheme
}

// ApplySepia replaces the dark native palette with warm light colors.
func ApplySepia(theme *ui.Theme) {
	accent := ui.Oklch(0.53, 0.09, 65)
	theme.Dark = false
	theme.Background = ui.Oklch(0.86, 0.025, 80)
	theme.Surface = ui.Oklch(0.82, 0.025, 80)
	theme.SurfaceHover = ui.Oklch(0.79, 0.028, 78)
	theme.SurfacePressed = ui.Oklch(0.75, 0.032, 77)
	theme.Border = ui.Oklch(0.77, 0.02, 78)
	theme.Text = ui.Oklch(0.24, 0.02, 70)
	theme.TextMuted = ui.Oklch(0.44, 0.022, 70)
	theme.Accent = accent
	theme.AccentHover = ui.Oklch(0.47, 0.095, 65)
	theme.AccentPressed = ui.Oklch(0.42, 0.095, 63)
	theme.AccentText = ui.Oklch(0.97, 0.012, 80)
	theme.Danger = ui.Oklch(0.58, 0.22, 27)
	theme.Warning = ui.Oklch(0.66, 0.13, 65)
	theme.Success = ui.Oklch(0.55, 0.12, 145)
	theme.Selection = accent.Alpha(0.22)
	theme.Focus = accent.Alpha(0.55)
	theme.Scrollbar = ui.Oklch(0.42, 0.02, 70).Alpha(0.32)
}

// FontScale converts the interface font preference to a scale.
func FontScale(fontSize string) float32 {
	switch fontSize {
	case app.FontSizeMedium:
		return 1.15
	case app.FontSizeLarge:
		return 1.3
	default:
		return 1
	}
}
