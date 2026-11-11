// Package widget contains reusable stateless native interface controls.
package widget

import (
	"github.com/egoist/mygo/ui"
)

// IconButtonColors contains the palette needed to draw an icon button.
type IconButtonColors struct {
	Rest              ui.Color
	Active            ui.Color
	Accent            ui.Color
	Hover             ui.Color
	HoverBackground   ui.Color
	PressedBackground ui.Color
	DisabledOpacity   float32
}

// ThemeIconButtonColors creates the default icon-button palette for a theme.
func ThemeIconButtonColors(theme *ui.Theme) IconButtonColors {
	return IconButtonColors{
		Rest:              theme.Text,
		Active:            theme.AccentText,
		Accent:            theme.Warning,
		Hover:             theme.AccentText,
		HoverBackground:   theme.SurfaceHover,
		PressedBackground: theme.SurfacePressed,
		DisabledOpacity:   0.4,
	}
}

// IconButtonOptions describes the geometry and state of an icon button.
type IconButtonOptions struct {
	Icon     *ui.SVG
	Label    string
	Size     float32
	IconSize float32
	Radius   float32
	Enabled  bool
	Active   bool
	Accent   bool
	Action   func()
}

// IconButton draws a themed icon button and runs Action only when enabled.
func IconButton(c *ui.Context, colors IconButtonColors, options IconButtonOptions) ui.Element {
	var button ui.Element
	button = ui.ButtonBase(c).
		Size(options.Size, options.Size).
		Radius(options.Radius).
		Cursor(ui.CursorPointer).
		Label(options.Label).
		Tooltip(options.Label).
		Disabled(!options.Enabled)

	hovered := button.Hovered() && options.Enabled
	pressed := button.Pressed() && options.Enabled
	iconColor := colors.Rest
	if options.Active {
		iconColor = colors.Active
	}
	if options.Accent {
		iconColor = colors.Accent
	}
	if hovered || pressed {
		iconColor = colors.Hover
	}

	button.Children(func() {
		ui.Icon(c, options.Icon).Size(options.IconSize, options.IconSize).TextColor(iconColor)
	})
	if pressed {
		button.Background(colors.PressedBackground)
	} else if hovered {
		button.Background(colors.HoverBackground)
	} else {
		button.Background(ui.Transparent)
	}
	if !options.Enabled {
		button.Opacity(colors.DisabledOpacity)
	}
	if button.Clicked() && options.Enabled {
		options.Action()
	}
	return button
}
