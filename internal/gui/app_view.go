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
	confirmations *Confirmations
	sidebar       *FeedSidebar
}

// NewAppView creates the main window's native layout.
func NewAppView(
	statusModel *appstatus.Model,
	onAbout func(),
	notifications *Notifications,
	confirmations *Confirmations,
	sidebar *FeedSidebar,
) *AppView {
	return &AppView{
		status:        NewStatusBar(statusModel, onAbout),
		notifications: notifications,
		confirmations: confirmations,
		sidebar:       sidebar,
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
		} else if a.sidebar != nil && !a.sidebar.isCollapsed() {
			var split ui.Element
			ui.Themed(c, a.sidebar.SidebarTheme(c)).FillWidth().Grow(1).Children(func() {
				split = ui.SplitPercent(c, a.sidebar.splitPercent(), func() {
					a.sidebar.View(c)
				}, func() {
					ui.WebView(c, a.page).Grow(1)
				}).Fill()
			})
			if split.Changed() {
				a.sidebar.savePanePercent()
			}
		} else {
			ui.WebView(c, a.page).Grow(1)
			if a.sidebar != nil {
				a.sidebar.drawOverlay(c)
			}
		}
		a.status.View(c)
	})
	if a.notifications != nil {
		a.notifications.draw(c)
	}
	if a.confirmations != nil {
		a.confirmations.draw(c)
	}
}
