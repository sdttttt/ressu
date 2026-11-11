package dialog

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
)

// ErrNoConfirmationSurface reports that no window can render confirmations.
var ErrNoConfirmationSurface = errors.New("no confirmation surface")

// ErrInvalidConfirmation reports an empty confirmation prompt.
var ErrInvalidConfirmation = errors.New("confirmation prompt is empty")

// ConfirmationRequest describes one destructive-action prompt.
type ConfirmationRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ConfirmLabel string `json:"confirmLabel"`
	CancelLabel  string `json:"cancelLabel"`
	Destructive  bool   `json:"destructive"`
}

// confirmationResult carries the answer to a queued request.
type confirmationResult struct {
	accepted bool
	err      error
}

// queuedConfirmation is one pending confirmation and its completion signal.
type queuedConfirmation struct {
	ConfirmationRequest
	open      bool
	completed bool
	done      chan struct{}
	result    confirmationResult
}

// Confirmations renders GPU-native prompts inside the main window.
type Confirmations struct {
	mu         sync.Mutex
	window     *mygo.Window
	appearance func() (theme, fontSize string)
	queued     []*queuedConfirmation
	active     *queuedConfirmation
}

// NewConfirmations creates the in-app confirmation host.
func NewConfirmations(appearance func() (theme, fontSize string)) *Confirmations {
	return &Confirmations{appearance: appearance}
}

// Attach connects confirmation rendering to the main window.
func (c *Confirmations) Attach(window *mygo.Window) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.window = window
}

// Confirm shows a prompt and reports whether the user accepted it.
func (c *Confirmations) Confirm(
	ctx context.Context,
	parent *mygo.Window,
	request ConfirmationRequest,
) (bool, error) {
	request.Title = strings.TrimSpace(request.Title)
	request.Description = strings.TrimSpace(request.Description)
	if request.Title == "" || request.Description == "" {
		return false, ErrInvalidConfirmation
	}
	if request.ConfirmLabel == "" {
		request.ConfirmLabel = "确认"
	}
	if request.CancelLabel == "" {
		request.CancelLabel = "取消"
	}

	c.mu.Lock()
	surface := c.window
	c.mu.Unlock()
	if surface == nil || surface.IsDestroyed() {
		return false, ErrNoConfirmationSurface
	}

	confirmation := &queuedConfirmation{ConfirmationRequest: request, done: make(chan struct{})}
	c.enqueue(confirmation)
	off := surface.OnClosed(func() {
		c.complete(confirmation, false, context.Canceled)
	})
	c.invalidate()

	select {
	case <-ctx.Done():
		if c.cancel(confirmation) {
			off()
			c.invalidate()
		}
		return false, ctx.Err()
	case <-confirmation.done:
		off()
		return confirmation.result.accepted, confirmation.result.err
	}
}

// enqueue adds one prompt without moving it to the UI thread.
func (c *Confirmations) enqueue(confirmation *queuedConfirmation) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.active == nil && len(c.queued) == 0 {
		c.active = confirmation
		c.active.open = true
		return
	}
	c.queued = append(c.queued, confirmation)
}

// cancel removes a prompt before the user answers it.
func (c *Confirmations) cancel(confirmation *queuedConfirmation) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if confirmation.completed {
		return false
	}
	if c.active == confirmation {
		c.activateNextLocked()
	} else {
		for i, item := range c.queued {
			if item == confirmation {
				c.queued = append(c.queued[:i], c.queued[i+1:]...)
				break
			}
		}
	}
	c.completeLocked(confirmation, false, context.Canceled)
	return true
}

// complete records an answer and wakes the waiting caller.
func (c *Confirmations) complete(confirmation *queuedConfirmation, accepted bool, err error) {
	c.mu.Lock()
	if !c.removeLocked(confirmation) {
		c.mu.Unlock()
		return
	}
	c.completeLocked(confirmation, accepted, err)
	c.mu.Unlock()
	c.invalidate()
}

// removeLocked removes an active or queued prompt once.
func (c *Confirmations) removeLocked(confirmation *queuedConfirmation) bool {
	if confirmation.completed {
		return false
	}
	if c.active == confirmation {
		c.activateNextLocked()
	} else {
		found := false
		for i, item := range c.queued {
			if item == confirmation {
				c.queued = append(c.queued[:i], c.queued[i+1:]...)
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// completeLocked publishes an answer and closes its completion signal.
func (c *Confirmations) completeLocked(confirmation *queuedConfirmation, accepted bool, err error) {
	confirmation.completed = true
	confirmation.open = false
	confirmation.result = confirmationResult{accepted: accepted, err: err}
	close(confirmation.done)
}

// activateNextLocked promotes the next queued prompt outside the UI frame.
func (c *Confirmations) activateNextLocked() {
	if len(c.queued) == 0 {
		c.active = nil
		return
	}
	c.active = c.queued[0]
	c.queued = c.queued[1:]
	c.active.open = true
}

// invalidate schedules the next confirmation frame.
func (c *Confirmations) invalidate() {
	c.mu.Lock()
	window := c.window
	c.mu.Unlock()
	if window != nil && !window.IsDestroyed() {
		window.Invalidate()
	}
}

// Draw renders the active confirmation prompt.
func (c *Confirmations) Draw(uiContext *ui.Context) {
	c.mu.Lock()
	confirmation := c.active
	c.mu.Unlock()
	if confirmation == nil {
		return
	}
	c.drawConfirmation(uiContext, confirmation)
}

// drawConfirmation renders one modal and completes its pending request.
func (c *Confirmations) drawConfirmation(uiContext *ui.Context, confirmation *queuedConfirmation) {
	appearanceName, fontSize := c.appearance()
	theme := guitheme.New(uiContext, appearanceName, fontSize)

	ui.Themed(uiContext, theme).Children(func() {
		modal := ui.Modal(uiContext, &confirmation.open, func() {
			ui.Column(uiContext).Gap(14).MaxWidth(390).Children(func() {
				ui.Column(uiContext).Gap(7).Children(func() {
					ui.Text(uiContext, confirmation.Title).
						FontSize(17).
						Bold().
						Role(ui.RoleHeading).
						Level(2)
					ui.Text(uiContext, confirmation.Description).TextColor(theme.TextMuted)
				})
				ui.Row(uiContext).Gap(9).Justify(ui.End).Children(func() {
					if ui.Button(uiContext, confirmation.CancelLabel).Clicked() {
						c.complete(confirmation, false, nil)
					}
					button := ui.PrimaryButton(uiContext, confirmation.ConfirmLabel)
					if confirmation.Destructive {
						button.Background(theme.Danger).TextColor(ui.RGBA(250, 250, 250, 0.98))
					}
					if button.Clicked() {
						c.complete(confirmation, true, nil)
					}
				})
			})
		})
		modal.Role(ui.RoleAlertDialog).
			Label(confirmation.Title).
			Description(confirmation.Description)
		if modal.Dismissed() {
			c.complete(confirmation, false, nil)
		}
	})
}
