//go:build !darwin

package main

import (
	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/internal/action"
)

// setMenu removes the native menu bar, so no menu row sits above the page.
func setMenu(_ *action.AppActions) {
	mygo.App.SetMenu(nil)
}
