package main

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/internal/gui"
)

// appView places the reader page above the native status bar.
type appView struct {
	page   *mygo.WebView
	status *gui.StatusBar
}

// view builds the native layout around the reader page.
func (a *appView) view(c *ui.Context) {
	theme := c.Theme()
	ui.Column(c).Fill().Background(theme.Background).Children(func() {
		if a.page == nil {
			ui.Box(c).Grow(1).Background(theme.Background)
			return
		}
		ui.WebView(c, a.page).Grow(1)
		a.status.View(c)
	})
}
