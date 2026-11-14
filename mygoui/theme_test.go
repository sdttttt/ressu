package mygoui

import (
	"math"
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
	appsidebar "github.com/sdttttt/ressu/mygoui/sidebar"
	guitheme "github.com/sdttttt/ressu/mygoui/theme"
)

func TestComponentThemesIgnoreInheritedTextScale(t *testing.T) {
	tests := []struct {
		name     string
		fontSize string
		want     float32
	}{
		{"small", app.FontSizeSmall, 13},
		{"medium", app.FontSizeMedium, 14.95},
		{"large", app.FontSizeLarge, 16.9},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sidebarSize, postListSize float32
			view := func(c *ui.Context) {
				outer := appsidebar.Theme(c, app.ThemeSystem, test.fontSize)
				ui.Themed(c, outer).Children(func() {
					sidebarSize = appsidebar.Theme(c, app.ThemeSystem, test.fontSize).FontSize
					postListSize = guitheme.PostTheme(c, app.ThemeSystem, test.fontSize).FontSize
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
