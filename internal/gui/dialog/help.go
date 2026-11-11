package dialog

import (
	"sync/atomic"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
)

// helpShortcut describes one application keyboard action.
type helpShortcut struct {
	key         string
	description string
}

var helpShortcuts = []helpShortcut{
	{key: "Space", description: "向下滚动正文，到底时读下一篇"},
	{key: "m", description: "切换已读 / 未读"},
	{key: "v", description: "在浏览器中打开原文"},
	{key: "h", description: "上一个订阅源"},
	{key: "l", description: "下一个订阅源"},
	{key: "?", description: "显示快捷键帮助"},
}

// Help renders the native keyboard shortcut modal.
type Help struct {
	open       bool
	requested  atomic.Bool
	visible    atomic.Bool
	window     *mygo.Window
	appearance func() (theme, fontSize string)
}

// NewHelp creates the in-app keyboard shortcut host.
func NewHelp(appearance func() (theme, fontSize string)) *Help {
	return &Help{appearance: appearance}
}

// Attach connects the modal to the window that renders it.
func (h *Help) Attach(window *mygo.Window) {
	if h != nil {
		h.window = window
	}
}

// Show opens the keyboard shortcut modal.
func (h *Help) Show() {
	if h != nil {
		h.requested.Store(true)
		h.visible.Store(true)
		h.invalidate()
	}
}

// Visible reports whether the keyboard shortcut modal is open.
func (h *Help) Visible() bool {
	return h != nil && h.visible.Load()
}

// Draw renders the help modal when it is open.
func (h *Help) Draw(uiContext *ui.Context) {
	if h == nil {
		return
	}
	if h.requested.Swap(false) {
		h.open = true
		h.visible.Store(true)
	}
	if !h.open {
		return
	}

	appearanceName, fontSize := h.appearance()
	theme := guitheme.New(uiContext, appearanceName, fontSize)
	ui.Themed(uiContext, theme).Children(func() {
		modal := ui.Modal(uiContext, &h.open, func() {
			ui.Column(uiContext).MaxWidth(390).Gap(12).Children(func() {
				ui.Text(uiContext, "快捷键").FontSize(17).Bold().
					Role(ui.RoleHeading).Level(2)
				for _, shortcut := range helpShortcuts {
					drawHelpShortcut(uiContext, theme, shortcut)
				}
			})
		})
		modal.Role(ui.RoleDialog).Label("快捷键")
		if modal.Dismissed() {
			h.open = false
			h.visible.Store(false)
		}
	})
}

func (h *Help) invalidate() {
	if h.window != nil && !h.window.IsDestroyed() {
		h.window.Invalidate()
	}
}

// drawHelpShortcut renders one keyboard action row.
func drawHelpShortcut(uiContext *ui.Context, theme *ui.Theme, shortcut helpShortcut) {
	ui.Row(uiContext).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Box(uiContext).MinWidth(54).Padding(4, 8).
			Radius(6).Background(theme.SurfaceHover).Border(1, theme.Border).
			AlignItems(ui.Center).Children(func() {
			ui.Text(uiContext, shortcut.key).SingleLine()
		})
		ui.Text(uiContext, shortcut.description).
			Grow(1).TextColor(theme.TextMuted)
	})
}
