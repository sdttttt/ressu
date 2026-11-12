package service

import (
	"context"
	"testing"
	"time"
)

func TestForegroundRefresherRefreshesAfterReturning(t *testing.T) {
	refreshes := make(chan struct{}, 4)
	refresher := NewForegroundRefresher(func(context.Context) {
		refreshes <- struct{}{}
	})

	refresher.BecomeActive()
	select {
	case <-refreshes:
		t.Fatal("the first activation should not refresh")
	case <-time.After(10 * time.Millisecond):
	}

	refresher.ResignActive()
	refresher.BecomeActive()
	waitForRefresh(t, refreshes)

	refresher.BecomeActive()
	select {
	case <-refreshes:
		t.Fatal("staying active should not refresh again")
	case <-time.After(10 * time.Millisecond):
	}
}

func TestForegroundRefresherUsesWindowFocusFallback(t *testing.T) {
	refreshes := make(chan struct{}, 4)
	refresher := NewForegroundRefresher(func(context.Context) {
		refreshes <- struct{}{}
	})

	refresher.WindowBlurred(true)
	refresher.WindowFocused(true)
	select {
	case <-refreshes:
		t.Fatal("an application-internal focus change should not refresh")
	case <-time.After(10 * time.Millisecond):
	}

	refresher.WindowBlurred(false)
	refresher.WindowFocused(true)
	waitForRefresh(t, refreshes)
}

func waitForRefresh(t *testing.T, refreshes <-chan struct{}) {
	t.Helper()
	select {
	case <-refreshes:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("the app did not refresh after returning to the foreground")
	}
}
