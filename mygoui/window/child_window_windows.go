//go:build windows

package window

import "github.com/egoist/mygo"

// positionChildWindowOptions places a child window over its parent before
// MyGO creates it. MyGO centers a window without X and Y on the primary
// display on Windows, even when the window has a parent on another display.
func positionChildWindowOptions(parent *mygo.Window, options *mygo.WindowOptions) {
	if parent == nil || parent.IsDestroyed() {
		return
	}

	parentBounds := parent.Bounds()
	display := mygo.Screen.DisplayMatching(parentBounds)
	if display.WorkArea.Width <= 0 || display.WorkArea.Height <= 0 {
		return
	}

	width, height := options.Width, options.Height
	if width <= 0 {
		width = 800
	}
	if height <= 0 {
		height = 600
	}
	options.X, options.Y = centeredWindowPosition(
		parentBounds,
		mygo.Rectangle{Width: width, Height: height},
		display.WorkArea,
	)
}

// placeChildWindowOnParentDisplay corrects the position after creation.
// StateKey restores a window's previous bounds after MyGO applies the initial
// position, so a remembered settings window can still land on another display.
func placeChildWindowOnParentDisplay(parent, child *mygo.Window) {
	if parent == nil || child == nil || parent.IsDestroyed() || child.IsDestroyed() {
		return
	}

	parentBounds := parent.Bounds()
	parentDisplay := mygo.Screen.DisplayMatching(parentBounds)
	if parentDisplay.WorkArea.Width <= 0 || parentDisplay.WorkArea.Height <= 0 {
		return
	}
	childBounds := child.Bounds()
	if sameDisplay(parentDisplay, mygo.Screen.DisplayMatching(childBounds)) {
		return
	}

	child.SetPosition(centeredWindowPosition(parentBounds, childBounds, parentDisplay.WorkArea))
}

// centeredWindowPosition centers a window over its parent without letting it
// cross the parent display's work area.
func centeredWindowPosition(parent, window, workArea mygo.Rectangle) (int, int) {
	x := parent.X + (parent.Width-window.Width)/2
	y := parent.Y + (parent.Height-window.Height)/2
	x = max(workArea.X, min(x, workArea.X+workArea.Width-window.Width))
	y = max(workArea.Y, min(y, workArea.Y+workArea.Height-window.Height))
	return x, y
}

// sameDisplay reports whether two displays describe the same monitor.
func sameDisplay(a, b mygo.Display) bool {
	if a.ID != 0 && b.ID != 0 {
		return a.ID == b.ID
	}
	return a.Bounds == b.Bounds
}
