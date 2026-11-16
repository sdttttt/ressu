package mygoui

import (
	"testing"
	"time"
)

// newTestLifecycle builds a lifecycle with counting stand-ins for the
// platform hooks so tests can drive the state machine without a window.
func newTestLifecycle() (*pageLifecycle, *int, *int) {
	var suspends, resumes int
	l := &pageLifecycle{
		delay:   time.Millisecond,
		suspend: func() { suspends++ },
		resume:  func() { resumes++ },
		runOnMain: func(fn func()) {
			fn()
		},
	}
	return l, &suspends, &resumes
}

func TestLifecycleSuspendsAfterMinimizeDelay(t *testing.T) {
	l, suspends, resumes := newTestLifecycle()

	l.minimized()
	time.Sleep(10 * time.Millisecond) // let the timer fire

	if *suspends != 1 {
		t.Errorf("suspends = %d, want 1", *suspends)
	}
	if *resumes != 0 {
		t.Errorf("resumes = %d", *resumes)
	}
}

func TestLifecycleRestoreCancelsPendingSuspend(t *testing.T) {
	l, suspends, _ := newTestLifecycle()

	l.minimized()
	l.restored()
	time.Sleep(10 * time.Millisecond)

	if *suspends != 0 {
		t.Errorf("suspends after restore = %d", *suspends)
	}
}

func TestLifecycleRestoreRecreatesSuspendedPage(t *testing.T) {
	l, suspends, resumes := newTestLifecycle()

	l.minimized()
	time.Sleep(10 * time.Millisecond)
	l.restored()

	if *suspends != 1 {
		t.Errorf("suspends = %d, want 1", *suspends)
	}
	if *resumes != 1 {
		t.Errorf("resumes = %d, want 1", *resumes)
	}
}

func TestLifecycleRestoreWithoutSuspendIsNoop(t *testing.T) {
	l, suspends, resumes := newTestLifecycle()

	l.restored()

	if *suspends != 0 || *resumes != 0 {
		t.Errorf("suspends = %d, resumes = %d", *suspends, *resumes)
	}
}

func TestLifecycleRepeatedMinimizeWhilePending(t *testing.T) {
	l, suspends, _ := newTestLifecycle()

	l.minimized()
	l.minimized()
	time.Sleep(10 * time.Millisecond)

	if *suspends != 1 {
		t.Errorf("suspends = %d, want 1", *suspends)
	}
}
