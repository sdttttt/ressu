package theme

import (
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/gui/widget"
)

// postTextSize matches the article list and header's 13px text.
const postTextSize = 13

// nativeTextSize is MyGO's default text size before system text scaling.
var nativeTextSize = ui.LightTheme().FontSize

// PostTheme maps app appearance preferences to the article list palette.
func PostTheme(c *ui.Context, appearance, fontSize string) *ui.Theme {
	theme := New(c, appearance, fontSize)
	if nativeTextSize > 0 {
		theme.FontSize *= postTextSize / nativeTextSize
	}
	palette := NewPostPalette(*theme)
	theme.Background = palette.Background
	theme.Surface = ui.Transparent
	theme.SurfaceHover = palette.Accent
	theme.SurfacePressed = palette.Accent
	theme.Text = palette.Foreground
	theme.TextMuted = palette.TextMuted
	theme.Border = palette.Border
	theme.Scrollbar = palette.TextMuted.Alpha(0.4)
	return theme
}

// PostPalette contains the article list and header colors.
type PostPalette struct {
	Background          ui.Color
	Foreground          ui.Color
	TextMuted           ui.Color
	Secondary           ui.Color
	Muted               ui.Color
	SecondaryForeground ui.Color
	Accent              ui.Color
	AccentForeground    ui.Color
	Border              ui.Color
	Primary             ui.Color
	Warning             ui.Color
}

// NewPostPalette branches the article palette by appearance.
func NewPostPalette(base ui.Theme) PostPalette {
	if base.Dark {
		return PostPalette{
			Background:          ui.Oklch(0.17, 0.005, 285),
			Foreground:          ui.Oklch(0.93, 0.002, 285),
			TextMuted:           ui.Oklch(0.68, 0.01, 285),
			Secondary:           ui.Oklch(0.25, 0.006, 285),
			Muted:               ui.Oklch(0.24, 0.006, 285),
			SecondaryForeground: ui.Oklch(0.93, 0.002, 285),
			Accent:              ui.Oklch(0.27, 0.008, 285),
			AccentForeground:    ui.Oklch(0.95, 0.002, 285),
			Border:              ui.Oklch(0.30, 0.006, 285),
			Primary:             base.Accent,
			Warning:             ui.Oklch(0.80, 0.14, 80),
		}
	}
	if base.Background == ui.Oklch(0.86, 0.025, 80) {
		return PostPalette{
			Background:          ui.Oklch(0.86, 0.025, 80),
			Foreground:          ui.Oklch(0.24, 0.02, 70),
			TextMuted:           ui.Oklch(0.44, 0.022, 70),
			Secondary:           ui.Oklch(0.82, 0.025, 80),
			Muted:               ui.Oklch(0.82, 0.025, 80),
			SecondaryForeground: ui.Oklch(0.28, 0.02, 70),
			Accent:              ui.Oklch(0.80, 0.03, 78),
			AccentForeground:    ui.Oklch(0.26, 0.02, 70),
			Border:              ui.Oklch(0.77, 0.02, 78),
			Primary:             base.Accent,
			Warning:             ui.Oklch(0.66, 0.13, 65),
		}
	}
	return PostPalette{
		Background:          ui.Oklch(0.995, 0, 285),
		Foreground:          ui.Oklch(0.22, 0.006, 285),
		TextMuted:           ui.Oklch(0.53, 0.01, 285),
		Secondary:           ui.Oklch(0.96, 0.002, 285),
		Muted:               ui.Oklch(0.96, 0.002, 285),
		SecondaryForeground: ui.Oklch(0.28, 0.006, 285),
		Accent:              ui.Oklch(0.94, 0.004, 285),
		AccentForeground:    ui.Oklch(0.25, 0.006, 285),
		Border:              ui.Oklch(0.90, 0.003, 285),
		Primary:             base.Accent,
		Warning:             ui.Oklch(0.78, 0.15, 78),
	}
}

// PostIconButtonColors maps the article palette to icon button states.
func PostIconButtonColors(palette PostPalette) widget.IconButtonColors {
	return widget.IconButtonColors{
		Rest:              palette.Foreground,
		Active:            palette.Primary,
		Accent:            palette.Warning,
		Hover:             palette.AccentForeground,
		HoverBackground:   palette.Accent,
		PressedBackground: palette.Accent,
		DisabledOpacity:   0.4,
	}
}
