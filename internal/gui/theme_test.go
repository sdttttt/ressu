package gui

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/store"
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

func TestComponentThemesIgnoreInheritedTextScale(t *testing.T) {
	tests := []struct {
		name     string
		fontSize string
		want     float32
	}{
		{"small", store.FontSizeSmall, 13},
		{"medium", store.FontSizeMedium, 14.95},
		{"large", store.FontSizeLarge, 16.9},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sidebarSize, postListSize float32
			view := func(c *ui.Context) {
				outer := feedSidebarTheme(c, store.ThemeSystem, test.fontSize)
				ui.Themed(c, outer).Children(func() {
					sidebarSize = feedSidebarTheme(c, store.ThemeSystem, test.fontSize).FontSize
					postListSize = postListTheme(c, store.ThemeSystem, test.fontSize).FontSize
				})
			}

			tester := ui.NewTester(view, 100, 100)
			tester.Frame()

			if math.Abs(float64(sidebarSize-test.want)) > 0.001 {
				t.Errorf("nested sidebar font size = %v, want %v", sidebarSize, test.want)
			}
			if math.Abs(float64(postListSize-test.want)) > 0.001 {
				t.Errorf("nested article list font size = %v, want %v", postListSize, test.want)
			}
		})
	}
}
