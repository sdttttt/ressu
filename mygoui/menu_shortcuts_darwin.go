//go:build darwin

package mygoui

import (
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/app"
)

// installMenuShortcuts is a no-op on macOS: the system menu bar owns the
// CmdOrCtrl+N/R/, accelerators there.
func installMenuShortcuts(_ *ui.Context, _ *app.App) {}
