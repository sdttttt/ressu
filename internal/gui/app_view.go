package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appdialog "github.com/sdttttt/ressu/internal/gui/dialog"
	apppostheader "github.com/sdttttt/ressu/internal/gui/postheader"
	apppostlist "github.com/sdttttt/ressu/internal/gui/postlist"
	appsidebar "github.com/sdttttt/ressu/internal/gui/sidebar"
	appstatus "github.com/sdttttt/ressu/internal/status"
)

// AppView places the reader page above the native status bar.
type AppView struct {
	page          *mygo.WebView
	status        *StatusBar
	notifications *appdialog.Notifications
	confirmations *appdialog.Confirmations
	help          *appdialog.Help
	sidebar       *appsidebar.Sidebar
	postHeader    *apppostheader.Header
	postList      *apppostlist.List
}

// NewAppView creates the main window's native layout.
func NewAppView(
	statusModel *appstatus.Model,
	onAbout func(),
	notifications *appdialog.Notifications,
	confirmations *appdialog.Confirmations,
	help *appdialog.Help,
	sidebar *appsidebar.Sidebar,
	postHeader *apppostheader.Header,
	postList *apppostlist.List,
) *AppView {
	return &AppView{
		status:        NewStatusBar(statusModel, onAbout),
		notifications: notifications,
		confirmations: confirmations,
		help:          help,
		sidebar:       sidebar,
		postHeader:    postHeader,
		postList:      postList,
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
		} else {
			drawContent := func() {
				if a.postList == nil {
					ui.WebView(c, a.page).Grow(1)
					return
				}

				ui.Themed(c, a.postList.Theme(c)).Fill().Children(func() {
					var articleSplit ui.Element
					articleSplit = ui.SplitPercent(c, a.postList.SplitPercent(), func() {
						a.postList.View(c)
					}, func() {
						ui.Column(c).Fill().Children(func() {
							if a.postHeader != nil {
								a.postHeader.View(c)
							}
							ui.WebView(c, a.page).Grow(1)
						})
					}).Fill()
					if articleSplit.Changed() {
						a.postList.SavePanePercent()
					}
				})
			}

			if a.sidebar != nil && !a.sidebar.IsCollapsed() {
				var split ui.Element
				ui.Themed(c, a.sidebar.Theme(c)).FillWidth().Grow(1).Children(func() {
					split = ui.SplitPercent(c, a.sidebar.SplitPercent(), func() {
						a.sidebar.View(c)
					}, drawContent).Fill()
				})
				if split.Changed() {
					a.sidebar.SavePanePercent()
				}
			} else {
				drawContent()
				if a.sidebar != nil {
					a.sidebar.DrawOverlay(c)
				}
			}
		}
		a.status.View(c)
	})
	a.installShortcuts(c)
	if a.notifications != nil {
		a.notifications.Draw(c)
	}
	if a.confirmations != nil {
		a.confirmations.Draw(c)
	}
	if a.help != nil {
		a.help.Draw(c)
	}
}

// installShortcuts wires cross-surface navigation and the help modal.
func (a *AppView) installShortcuts(c *ui.Context) {
	if c.FocusTakesText() {
		return
	}
	if a.help != nil {
		c.OnWindowShortcut(ui.Shift, ui.KeySlash, a.help.Show)
	}
	if a.sidebar != nil {
		c.OnWindowShortcut(0, ui.KeyH, func() { a.sidebar.SelectNeighbour(-1) })
		c.OnWindowShortcut(0, ui.KeyL, func() { a.sidebar.SelectNeighbour(1) })
	}
}
