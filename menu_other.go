//go:build !darwin

package main

import (
	"github.com/egoist/mygo"

	"github.com/sdttttt/ressu/core/app"
)

// setMenu removes the native menu bar, so no menu row sits above the page.
func setMenu(_ *app.AppActions) {
	mygo.App.SetMenu(nil)
}
