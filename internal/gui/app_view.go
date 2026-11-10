package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appstatus "github.com/sdttttt/ressu/internal/status"
)

// AppView places the reader page above the native status bar.
type AppView struct {
	page          *mygo.WebView
	status        *StatusBar
	notifications *Notifications
}

// NewAppView creates the main window's native layout.
func NewAppView(
	statusModel *appstatus.Model,
	onAbout func(),
	notifications *Notifications,
) *AppView {
	return &AppView{
		status:        NewStatusBar(statusModel, onAbout),
		notifications: notifications,
	}
}

// SetPage attaches the reader page shown above the status bar.
func (a *AppView) SetPage(page *mygo.WebView) {
	a.page = page
}

// View builds the native layout around the reader page.
func (a *AppView) View(c *ui.Context) {
	theme := c.Theme()
	ui.Column(c).Fill().Background(theme.Background).Children(func() {
		if a.page == nil {
			ui.Box(c).Grow(1).Background(theme.Background)
			return
		}
		ui.WebView(c, a.page).Grow(1)
		a.status.View(c)
	})
	if a.notifications != nil {
		a.notifications.draw(c)
	}
}
