//go:build !windows

package gui

import "github.com/egoist/mygo"

// positionChildWindowOptions leaves child window placement to MyGO outside
// Windows.
func positionChildWindowOptions(*Window, *mygo.WindowOptions) {}

// placeChildWindowOnParentDisplay leaves child window placement to MyGO
// outside Windows.
func placeChildWindowOnParentDisplay(*Window, *Window) {}
