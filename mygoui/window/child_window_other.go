//go:build !windows

package window

import "github.com/egoist/mygo"

// positionChildWindowOptions leaves child window placement to MyGO outside
// Windows.
func positionChildWindowOptions(*mygo.Window, *mygo.WindowOptions) {}

// placeChildWindowOnParentDisplay leaves child window placement to MyGO
// outside Windows.
func placeChildWindowOnParentDisplay(*mygo.Window, *mygo.Window) {}
