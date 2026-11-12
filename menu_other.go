//go:build !darwin

package main

import (
	"github.com/egoist/mygo"

	appevent "github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/internal/action"
)

// setMenu removes the native menu bar, so no menu row sits above the page.
func setMenu(_ *appevent.Hub, _ *action.AppActions) {
	mygo.App.SetMenu(nil)
}
