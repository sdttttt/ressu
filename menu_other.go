//go:build !darwin

package main

import "github.com/egoist/mygo"

// setMenu removes the native menu bar, so no menu row sits above the page.
func setMenu() {
	mygo.App.SetMenu(nil)
}
