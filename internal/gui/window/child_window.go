// Package window creates native child windows with platform placement fixes.
package window

import "github.com/egoist/mygo"

// New creates a child window and applies the Windows-only placement
// workaround around MyGO's normal window creation.
func New(parent *mygo.Window, options mygo.WindowOptions) *mygo.Window {
	positionChildWindowOptions(parent, &options)
	window := mygo.NewWindow(options)
	placeChildWindowOnParentDisplay(parent, window)
	return window
}
