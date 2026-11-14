package widget_test

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/mygoui/icon"
	"github.com/sdttttt/ressu/mygoui/widget"
)

func TestIconButtonRunsEnabledAction(t *testing.T) {
	clicked := make(chan struct{}, 1)
	tester := ui.NewTester(func(c *ui.Context) {
		theme := ui.LightTheme()
		widget.IconButton(c, widget.ThemeIconButtonColors(theme), widget.IconButtonOptions{
			Icon:     icon.Search,
			Label:    "Test action",
			Size:     28,
			IconSize: 16,
			Radius:   6,
			Enabled:  true,
			Action: func() {
				clicked <- struct{}{}
			},
		})
	}, 100, 100)

	tester.Frame()
	if err := tester.Click("Test action"); err != nil {
		t.Fatalf("click icon button: %v", err)
	}
	select {
	case <-clicked:
	case <-time.After(time.Second):
		t.Fatal("clicking the icon button did not run its action")
	}
}
