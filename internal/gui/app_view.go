package gui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	appdialog "github.com/sdttttt/ressu/internal/gui/dialog"
	appstatus "github.com/sdttttt/ressu/internal/status"
)

// AppView places the reader page above the native status bar.
type AppView struct {
	page          *mygo.WebView
	status        *StatusBar
	notifications *appdialog.Notifications
	confirmations *appdialog.Confirmations
	sidebar       *FeedSidebar
	postHeader    *PostHeader
	postList      *PostList
}

// NewAppView creates the main window's native layout.
func NewAppView(
	statusModel *appstatus.Model,
	onAbout func(),
	notifications *appdialog.Notifications,
	confirmations *appdialog.Confirmations,
	sidebar *FeedSidebar,
	postHeader *PostHeader,
	postList *PostList,
) *AppView {
	return &AppView{
		status:        NewStatusBar(statusModel, onAbout),
		notifications: notifications,
		confirmations: confirmations,
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

				ui.Themed(c, a.postList.PostListTheme(c)).Fill().Children(func() {
					var articleSplit ui.Element
					articleSplit = ui.SplitPercent(c, a.postList.splitPercent(), func() {
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
						a.postList.savePanePercent()
					}
				})
			}

			if a.sidebar != nil && !a.sidebar.isCollapsed() {
				var split ui.Element
				ui.Themed(c, a.sidebar.SidebarTheme(c)).FillWidth().Grow(1).Children(func() {
					split = ui.SplitPercent(c, a.sidebar.splitPercent(), func() {
						a.sidebar.View(c)
					}, drawContent).Fill()
				})
				if split.Changed() {
					a.sidebar.savePanePercent()
				}
			} else {
				drawContent()
				if a.sidebar != nil {
					a.sidebar.drawOverlay(c)
				}
			}
		}
		a.status.View(c)
	})
	if a.notifications != nil {
		a.notifications.Draw(c)
	}
	if a.confirmations != nil {
		a.confirmations.Draw(c)
	}
}
