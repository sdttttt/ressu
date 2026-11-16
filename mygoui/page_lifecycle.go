package mygoui

import (
	"log/slog"
	"sync"
	"time"

	"github.com/egoist/mygo"
)

// pageSuspendDelay is how long the window must stay minimized before the
// reader page drops.
const pageSuspendDelay = 5 * time.Second

// pageLifecycle drops the reader web view after a delay once the main
// window is minimized, and recreates it when the window is restored, so the
// renderer processes release their memory. Recreation reloads the page,
// which pulls a fresh snapshot and its selected article again.
type pageLifecycle struct {
	delay     time.Duration
	suspend   func() // drops the current page; runs on the main thread
	resume    func() // recreates the page; runs on the main thread
	runOnMain func(func())

	mu    sync.Mutex
	state pageLifecycleState
	timer *time.Timer
}

// pageLifecycleState tracks the reader page across background suspensions.
type pageLifecycleState int

const (
	pageActive pageLifecycleState = iota
	pagePendingSuspend
	pageSuspended
)

// newPageLifecycle wires the reader page's suspension to the main window's
// minimize and restore signals.
func newPageLifecycle(win *mygo.Window, view *AppView) *pageLifecycle {
	l := &pageLifecycle{
		delay:     pageSuspendDelay,
		runOnMain: mygo.RunOnMain,
		suspend: func() {
			if win.IsDestroyed() {
				return
			}
			page := view.Page()
			view.SetPage(nil)
			win.Invalidate()
			if page != nil && !page.IsDestroyed() {
				page.Destroy()
			}
		},
		resume: func() {
			if win.IsDestroyed() {
				return
			}
			page, err := win.NewWebView(mygo.WebViewOptions{
				URL:             "/",
				BackgroundColor: "light-dark(#ffffff, #101012)",
			})
			if err != nil {
				slog.Warn("reader page recreate failed",
					"operation", "ui.pageResume", "error", err)
				return
			}
			view.SetPage(page)
			win.Invalidate()
		},
	}
	win.OnMinimize(l.minimized)
	win.OnRestore(l.restored)
	return l
}

// minimized schedules the suspension once the window goes to the taskbar.
func (l *pageLifecycle) minimized() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.state != pageActive {
		return
	}
	l.state = pagePendingSuspend
	l.timer = time.AfterFunc(l.delay, l.elapsed)
}

// restored cancels a pending suspension or recreates the dropped page.
func (l *pageLifecycle) restored() {
	l.mu.Lock()
	if l.state == pagePendingSuspend {
		l.stopTimerLocked()
		l.state = pageActive
		l.mu.Unlock()
		return
	}
	if l.state != pageSuspended {
		l.mu.Unlock()
		return
	}
	l.state = pageActive
	l.mu.Unlock()
	l.runOnMain(l.resume)
}

// elapsed drops the page once the minimize delay has passed.
func (l *pageLifecycle) elapsed() {
	l.mu.Lock()
	if l.state != pagePendingSuspend {
		l.mu.Unlock()
		return
	}
	l.state = pageSuspended
	l.mu.Unlock()
	l.runOnMain(l.suspend)
}

func (l *pageLifecycle) stopTimerLocked() {
	if l.timer != nil {
		l.timer.Stop()
		l.timer = nil
	}
}
