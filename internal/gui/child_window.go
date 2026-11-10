package gui

import "github.com/egoist/mygo"

// newChildWindow creates a child window and applies the Windows-only
// placement workaround around MyGO's normal window creation.
func newChildWindow(parent *Window, options mygo.WindowOptions) *Window {
	positionChildWindowOptions(parent, &options)
	window := mygo.NewWindow(options)
	placeChildWindowOnParentDisplay(parent, window)
	return window
}
