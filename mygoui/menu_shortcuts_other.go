//go:build !darwin

package mygoui

import (
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

// installMenuShortcuts registers the accelerators the macOS menu bar provides,
// since other platforms have no native menu row.
func installMenuShortcuts(c *ui.Context, actions *app.App) {
	if actions == nil || c.FocusTakesText() {
		return
	}
	c.OnWindowShortcut(ui.Ctrl, ui.KeyN, func() {
		actions.OpenAddFeed(app.NewWindowHandle(nil))
	})
	c.OnWindowShortcut(ui.Ctrl, ui.KeyR, func() { refreshAllWithNotice(actions) })
	c.OnWindowShortcut(ui.Ctrl, ui.KeyComma, func() {
		actions.OpenSettings(app.NewWindowHandle(nil))
	})
}
