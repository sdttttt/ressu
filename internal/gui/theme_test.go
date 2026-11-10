package gui

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestSepiaThemeUsesWarmInteractiveColors(t *testing.T) {
	theme := ui.DarkTheme()
	applySepiaTheme(theme)

	if theme.Dark {
		t.Fatal("sepia theme must use light widget behavior")
	}
	tests := []struct {
		name string
		got  ui.Color
		want ui.Color
	}{
		{"surface hover", theme.SurfaceHover, ui.Oklch(0.79, 0.028, 78)},
		{"surface pressed", theme.SurfacePressed, ui.Oklch(0.75, 0.032, 77)},
		{"accent", theme.Accent, ui.Oklch(0.53, 0.09, 65)},
		{"accent hover", theme.AccentHover, ui.Oklch(0.47, 0.095, 65)},
		{"accent pressed", theme.AccentPressed, ui.Oklch(0.42, 0.095, 63)},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %v, want %v", test.name, test.got, test.want)
		}
	}
}
