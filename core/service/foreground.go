package service

import (
	"context"
	"sync"

	"github.com/sdttttt/ressu/core/logging"
)

// ForegroundRefresher refreshes feeds once the app leaves and regains the
// foreground.
type ForegroundRefresher struct {
	refresh func(context.Context)

	mu       sync.Mutex
	inactive bool
}

// NewForegroundRefresher runs refresh when the app becomes foreground again.
func NewForegroundRefresher(refresh func(context.Context)) *ForegroundRefresher {
	return &ForegroundRefresher{refresh: refresh}
}

// ResignActive records that the app stopped being the foreground app.
func (r *ForegroundRefresher) ResignActive() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inactive = true
}

// BecomeActive refreshes when the app returns from the background.
func (r *ForegroundRefresher) BecomeActive() {
	r.mu.Lock()
	if !r.inactive {
		r.mu.Unlock()
		return
	}
	r.inactive = false
	refresh := r.refresh
	r.mu.Unlock()

	logging.Debug("foreground refresh triggered", "operation", "foreground.becomeActive")
	go refresh(context.Background())
}

// WindowBlurred applies the fallback signal of a window losing focus. active
// says whether the application itself remains active.
func (r *ForegroundRefresher) WindowBlurred(active bool) {
	if !active {
		r.ResignActive()
	}
}

// WindowFocused applies the fallback signal of a window gaining focus. active
// says whether the application itself is active.
func (r *ForegroundRefresher) WindowFocused(active bool) {
	if active {
		r.BecomeActive()
	}
}
