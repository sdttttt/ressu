package gui

import (
	"sync"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/sdttttt/ressu/internal/gui/assets"
	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
)

const (
	// aboutWidth is the content width of the About window.
	aboutWidth = 440

	// aboutHeight is the content height of the About window.
	aboutHeight = 360
)

var aboutWindow struct {
	sync.Mutex
	window *Window
}

// aboutAvatar decodes the embedded portrait once for every About window.
var aboutAvatar = sync.OnceValue(func() *ui.Bitmap {
	bitmap, err := ui.DecodeBitmap(assets.AboutAvatar)
	if err != nil {
		return nil
	}
	return bitmap
})

// AboutInfo is the metadata shown by the native About window.
type AboutInfo struct {
	Version   string
	BuildDate string
	Developer string
	Email     string
	Theme     string
	FontSize  string
}

// OpenAboutWindow opens the single native About window.
func OpenAboutWindow(parent *Window, info AboutInfo) {
	aboutWindow.Lock()
	if window := aboutWindow.window; window != nil && !window.IsDestroyed() {
		aboutWindow.Unlock()
		window.Show()
		window.Focus()
		return
	}
	aboutWindow.Unlock()

	view := &aboutView{info: info}
	window := newChildWindow(parent, mygo.WindowOptions{
		Title:           "关于 Ressu",
		Width:           aboutWidth,
		Height:          aboutHeight,
		UseContentSize:  true,
		DisableResize:   true,
		DisableMinimize: true,
		DisableMaximize: true,
		Parent:          parent,
		BackgroundColor: "light-dark(#ffffff, #101012)",
		Content:         ui.View(func(c *ui.Context) { view.draw(c) }),
	})
	aboutWindow.Lock()
	aboutWindow.window = window
	aboutWindow.Unlock()
	window.OnClosed(func() {
		aboutWindow.Lock()
		if aboutWindow.window == window {
			aboutWindow.window = nil
		}
		aboutWindow.Unlock()
	})
}

// aboutView draws the metadata in the About window.
type aboutView struct {
	info AboutInfo
}

// draw renders one frame of the About window.
func (a *aboutView) draw(c *ui.Context) {
	theme := guitheme.New(c, a.info.Theme, a.info.FontSize)

	ui.Themed(c, theme).Fill().Children(func() {
		ui.Column(c).Fill().Padding(24, 28).Gap(12).Children(func() {
			ui.Column(c).Gap(3).Children(func() {
				ui.Text(c, "Ressu").Bold().FontSize(26)
				ui.Textf(c, "版本 %s", a.info.Version).TextColor(theme.TextMuted)
			})
			ui.Box(c).Height(1).Shrink(0).Background(theme.Border)
			aboutDeveloper(c, theme, a.info.Developer)
			aboutField(c, theme, "构建日期", a.info.BuildDate)
			aboutField(c, theme, "邮箱", a.info.Email, "mailto:"+a.info.Email)
			ui.Text(c, "必开花繁盛，乐上加乐，而且欢呼。").TextColor(theme.TextMuted)
		})
	})
}

// aboutDeveloper renders the developer's portrait, label, and name.
func aboutDeveloper(c *ui.Context, theme *ui.Theme, developer string) {
	ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
		ui.Avatar(c, developer, aboutAvatar())
		ui.Column(c).Gap(2).Children(func() {
			ui.Text(c, "开发者").TextColor(theme.TextMuted)
			ui.Text(c, developer).Selectable()
		})
	})
}

// aboutField renders one label and value pair in the About window. A link
// is added when the value is also an address that can be opened.
func aboutField(c *ui.Context, theme *ui.Theme, label, value string, link ...string) {
	ui.Column(c).Gap(2).Children(func() {
		ui.Text(c, label).TextColor(theme.TextMuted)
		if len(link) > 0 && link[0] != "" && value != "" {
			ui.Link(c, value, link[0])
			return
		}
		ui.Text(c, value).Selectable()
	})
}
