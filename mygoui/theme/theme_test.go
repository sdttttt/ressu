package theme_test

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/mygoui/theme"
)

func TestApplySepiaUsesWarmInteractiveColors(t *testing.T) {
	native := ui.DarkTheme()
	theme.ApplySepia(native)

	if native.Dark {
		t.Fatal("sepia theme must use light widget behavior")
	}
	tests := []struct {
		name string
		got  ui.Color
		want ui.Color
	}{
		{"surface hover", native.SurfaceHover, ui.Oklch(0.79, 0.028, 78)},
		{"surface pressed", native.SurfacePressed, ui.Oklch(0.75, 0.032, 77)},
		{"accent", native.Accent, ui.Oklch(0.53, 0.09, 65)},
		{"accent hover", native.AccentHover, ui.Oklch(0.47, 0.095, 65)},
		{"accent pressed", native.AccentPressed, ui.Oklch(0.42, 0.095, 63)},
	}
	for _, test := range tests {
		if test.got != test.want {
			t.Errorf("%s = %v, want %v", test.name, test.got, test.want)
		}
	}
}
