// Package dialog renders reusable native confirmations and notifications.
package dialog

import (
	"strings"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
)

const maxPendingNotifications = 32

// Notification is a native toast requested outside the main view frame.
type Notification struct {
	Type        string
	Title       string
	Description string
}

// Notifications queues native toasts until the main window can show them.
type Notifications struct {
	mu         sync.Mutex
	window     *mygo.Window
	appearance func() (theme, fontSize string)
	pending    []Notification
}

// NewNotifications creates a notification queue.
func NewNotifications(appearance func() (theme, fontSize string)) *Notifications {
	return &Notifications{appearance: appearance}
}

// Attach connects the queue to the window that draws native toasts.
func (n *Notifications) Attach(window *mygo.Window) {
	n.mu.Lock()
	n.window = window
	pending := len(n.pending) > 0
	n.mu.Unlock()
	if pending {
		n.invalidate(window)
	}
}

// Show queues a notification and asks the main window to draw it.
func (n *Notifications) Show(notification Notification) {
	notification = normalizeNotification(notification)
	if notification.Title == "" {
		return
	}

	n.mu.Lock()
	if len(n.pending) >= maxPendingNotifications {
		n.pending = append(n.pending[:0], n.pending[1:]...)
	}
	n.pending = append(n.pending, notification)
	window := n.window
	n.mu.Unlock()
	n.invalidate(window)
}

func (n *Notifications) invalidate(window *mygo.Window) {
	if window != nil && !window.IsDestroyed() {
		window.Invalidate()
	}
}

func normalizeNotification(notification Notification) Notification {
	notification.Type = strings.ToLower(strings.TrimSpace(notification.Type))
	switch notification.Type {
	case "error", "warning", "info", "success":
	default:
		notification.Type = "info"
	}
	notification.Title = strings.TrimSpace(notification.Title)
	notification.Description = strings.TrimSpace(notification.Description)
	return notification
}

func (n *Notifications) Draw(c *ui.Context) {
	n.mu.Lock()
	pending := n.pending
	n.pending = nil
	n.mu.Unlock()

	for _, notification := range pending {
		c.AddToast(ui.Toast{
			Type:        notification.Type,
			Title:       notification.Title,
			Description: notification.Description,
		})
	}

	appearance := n.appearance
	if appearance == nil {
		appearance = func() (string, string) { return "", "" }
	}
	appearanceName, fontSize := appearance()
	theme := guitheme.New(c, appearanceName, fontSize)

	ui.ToastViewportBase(c, func(viewport ui.Element, toasts []ui.Toast) {
		viewport.Bottom(guitheme.StatusBarHeight(fontSize)).Padding(theme.Space(4)).
			AlignItems(ui.End).Gap(theme.Space(2))
		for _, toast := range toasts {
			drawNotification(c, theme, toast)
		}
	})
}

func drawNotification(c *ui.Context, theme *ui.Theme, toast ui.Toast) {
	part := ui.ToastBase(c, toast)
	root := part.Root.Row().AlignItems(ui.Stretch).Gap(theme.Space(2.5)).
		Padding(theme.Space(2.5), theme.Space(3.5)).Radius(theme.Space(2)).
		Background(theme.Surface).Border(1, theme.Border).MaxWidth(420).
		Shadow(0, 6, 20, 0, ui.RGBA(0, 0, 0, 0.18))

	applyToastFade(c, part, root)

	root.Children(func() {
		ui.Box(c).Width(theme.Space(0.75)).Shrink(0).Radius(theme.Space(0.5)).
			Background(notificationColor(theme, toast.Type))

		ui.Column(c).Shrink(1).Gap(theme.Space(0.75)).Children(func() {
			ui.Text(c, toast.Title).FontSize(theme.FontSize).FontWeight(600).
				TextColor(theme.Text)
			if toast.Description != "" {
				ui.Text(c, toast.Description).FontSize(theme.FontSize * 0.82).
					TextColor(theme.TextMuted)
			}
		})

		closeButton := part.CloseButton().Shrink(0).Label("关闭通知").
			Padding(theme.Space(0.5), theme.Space(1)).Radius(theme.Space(1)).
			Background(ui.Transparent)
		closeButton.Children(func() {
			ui.Text(c, "×").FontSize(theme.FontSize * 1.15).TextColor(theme.TextMuted)
		})
	})
}

func applyToastFade(c *ui.Context, part ui.ToastParts, root ui.Element) {
	const fade = 180 * time.Millisecond
	left := part.Left()
	if left <= 0 {
		return
	}
	if left <= fade {
		root.Opacity(float32(left) / float32(fade))
		c.AnimationFrame()
		return
	}
	c.After(left - fade)
}

func notificationColor(theme *ui.Theme, kind string) ui.Color {
	switch kind {
	case "error":
		return theme.Danger
	case "warning":
		return theme.Warning
	case "success":
		return theme.Success
	default:
		return theme.Accent
	}
}
