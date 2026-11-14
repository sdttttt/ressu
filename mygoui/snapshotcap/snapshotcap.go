// Package snapshotcap hosts the page window that captures article snapshots
// at a fixed card width, independent of the reader pane's width.
package snapshotcap

import (
	"sync"

	"github.com/egoist/mygo"

	appwindow "github.com/sdttttt/ressu/mygoui/window"
)

const (
	// captureWidth is the initial content width of the capture window.
	captureWidth = 860

	// captureHeight is the initial content height of the capture window.
	captureHeight = 900
)

var captureWindow struct {
	sync.Mutex
	window *mygo.Window
}

// Open opens the single snapshot capture window over parent. An existing
// window is shown and focused instead.
func Open(parent *mygo.Window) {
	captureWindow.Lock()
	if window := captureWindow.window; window != nil && !window.IsDestroyed() {
		captureWindow.Unlock()
		window.Show()
		window.Focus()
		return
	}
	captureWindow.Unlock()

	window := appwindow.New(parent, mygo.WindowOptions{
		Title:           "文章快照",
		URL:             "/snapshot.html",
		Width:           captureWidth,
		Height:          captureHeight,
		UseContentSize:  true,
		MinWidth:        800,
		MinHeight:       560,
		Parent:          parent,
		StateKey:        "snapshot-capture",
		BackgroundColor: "#ffffff",
	})

	captureWindow.Lock()
	captureWindow.window = window
	captureWindow.Unlock()
	window.OnClosed(func() {
		captureWindow.Lock()
		if captureWindow.window == window {
			captureWindow.window = nil
		}
		captureWindow.Unlock()
	})
}

// Close closes the snapshot capture window when it is the current one.
func Close(window *mygo.Window) {
	if window == nil || window.IsDestroyed() {
		return
	}
	captureWindow.Lock()
	current := captureWindow.window
	captureWindow.Unlock()
	if current == window {
		window.Close()
	}
}
