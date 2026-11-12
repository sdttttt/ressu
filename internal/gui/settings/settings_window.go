// Package settings renders the native application preferences window.
package settings

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/core/store"
	"github.com/sdttttt/ressu/internal/gui/format"
	guitheme "github.com/sdttttt/ressu/internal/gui/theme"
	appwindow "github.com/sdttttt/ressu/internal/gui/window"
)

const (
	// settingsWidth is the initial content width of the settings window.
	settingsWidth = 580

	// settingsHeight is the initial content height of the settings window.
	settingsHeight = 720
)

var settingsWindow struct {
	sync.Mutex
	window *mygo.Window
}

// Callbacks are the operations a native settings surface may invoke.
type Callbacks struct {
	Load         func() store.Prefs
	Save         func(store.Prefs) (store.Prefs, error)
	DataDir      func() string
	StorageUsage func() service.StorageInfo
	ExportOPML   func(parent *mygo.Window) (string, error)
	OpenDataDir  func() error
}

// Open opens the single native settings window.
func Open(parent *mygo.Window, callbacks Callbacks) {
	settingsWindow.Lock()
	if window := settingsWindow.window; window != nil && !window.IsDestroyed() {
		settingsWindow.Unlock()
		window.Show()
		window.Focus()
		return
	}
	settingsWindow.Unlock()

	view := &settingsView{callbacks: callbacks, prefs: callbacks.Load()}
	options := mygo.WindowOptions{
		Title:           "设置",
		Width:           settingsWidth,
		Height:          settingsHeight,
		UseContentSize:  true,
		MinWidth:        480,
		MinHeight:       560,
		Parent:          parent,
		StateKey:        "settings",
		BackgroundColor: "light-dark(#ffffff, #101012)",
		Content:         ui.View(func(c *ui.Context) { view.draw(c) }),
	}
	window := appwindow.New(parent, options)
	view.window = window

	settingsWindow.Lock()
	settingsWindow.window = window
	settingsWindow.Unlock()
	window.OnClosed(func() {
		settingsWindow.Lock()
		if settingsWindow.window == window {
			settingsWindow.window = nil
		}
		settingsWindow.Unlock()
	})
}

// settingsView keeps the editable preferences and the last save result.
type settingsView struct {
	callbacks Callbacks
	prefs     store.Prefs
	window    *mygo.Window
	err       string
	notice    string
	noticeAt  time.Time

	storage   service.StorageInfo
	storageAt time.Time
}

// draw renders one frame of the settings form.
func (s *settingsView) draw(c *ui.Context) {
	theme := guitheme.New(c, s.prefs.Theme, s.prefs.FontSize)
	changed := false
	s.refreshStorage(c)

	ui.Themed(c, theme).Fill().Children(func() {
		ui.Scroll(c).Grow(1).MinHeight(0).Padding(20, 24).Children(func() {
			ui.Form(c, func() {
				ui.Fieldset(c, "外观", func() {
					ui.Field(c, "主题", func() {
						themes := []string{
							store.ThemeSystem,
							store.ThemeLight,
							store.ThemeDark,
							store.ThemeSepia,
						}
						selected := indexOf(themes, s.prefs.Theme)
						if ui.Segmented(c, &selected, "跟随系统", "浅色", "深色", "护眼").
							Label("主题").Changed() {
							s.prefs.Theme = themes[selected]
							changed = true
						}
					})
					ui.Field(c, "界面字号", func() {
						fontSizes := []string{
							store.FontSizeSmall,
							store.FontSizeMedium,
							store.FontSizeLarge,
						}
						selected := indexOf(fontSizes, s.prefs.FontSize)
						if ui.Segmented(c, &selected, "小", "中", "大").
							Label("界面字号").Changed() {
							s.prefs.FontSize = fontSizes[selected]
							changed = true
						}
					})
				})

				ui.Fieldset(c, "阅读", func() {
					ui.Field(c, "文章排序依据", func() {
						sortFields := []string{
							store.PostSortPublished,
							store.PostSortFetched,
						}
						selected := indexOf(sortFields, s.prefs.PostSort)
						if ui.Segmented(c, &selected, "发布时间", "抓取时间").
							Label("文章排序依据").Changed() {
							s.prefs.PostSort = sortFields[selected]
							changed = true
						}
					}).Description("升序/降序按钮仍然可用，决定最新还是最旧在前。")
					ui.Field(c, "打开文章时标记为已读", func() {
						if ui.Switch(c, &s.prefs.MarkReadOnOpen).Changed() {
							changed = true
						}
					})
					ui.Field(c, "启动时自动刷新", func() {
						if ui.Switch(c, &s.prefs.RefreshOnStart).Changed() {
							changed = true
						}
					})
					ui.Field(c, "每个订阅源保留文章数", func() {
						value := float64(s.prefs.PostsPerFeed)
						if ui.NumberInput(c, &value, 50, 2000, 50).Changed() {
							s.prefs.PostsPerFeed = int(math.Round(value))
							changed = true
						}
					}).Description("范围 50–2000。")
				})

				ui.Fieldset(c, "远程资源", func() {
					ui.Field(c, "正文图片", func() {
						ui.RadioGroup(c, func() {
							if ui.Radio(c, &s.prefs.RemoteResources,
								store.ResourceAccessGoProxy, "由 Ressu 代理并缓存").Changed() {
								changed = true
							}
							if ui.Radio(c, &s.prefs.RemoteResources,
								store.ResourceAccessWebview, "由 WebView 直接加载").Changed() {
								changed = true
							}
						})
					}).Description("Ressu代理: 客户端主动抓取资源，可能因为请求指纹等问题被限制").Description("WebView: 直接访问源地址，可能会因CORS等策略被拦截").Description("(总之这个功能槽点很多，默认使用WebView模式)")
					ui.Field(c, "图片缓存上限（MB）", func() {
						value := float64(s.prefs.ImageCacheLimit)
						if ui.NumberInput(c, &value, 32, 512, 16).Changed() {
							s.prefs.ImageCacheLimit = int(math.Round(value))
							changed = true
						}
					}).Description(cacheUsageDescription(s.storage))
				})

				ui.Fieldset(c, "链接", func() {
					ui.Field(c, "打开文章链接", func() {
						ui.RadioGroup(c, func() {
							if ui.Radio(c, &s.prefs.OpenLinks,
								store.OpenLinksBrowser, "系统浏览器").Changed() {
								changed = true
							}
							if ui.Radio(c, &s.prefs.OpenLinks,
								store.OpenLinksWindow, "应用内窗口").Changed() {
								changed = true
							}
						})
					})
				})

				ui.Fieldset(c, "数据", func() {
					ui.Column(c).Gap(8).Children(func() {
						ui.Text(c, "数据目录").TextColor(theme.TextMuted)
						ui.Text(c, s.callbacks.DataDir()).Selectable().Font("monospace")
						ui.Text(c, dbSizeDescription(s.storage)).TextColor(theme.TextMuted)
						ui.Row(c).Gap(8).Children(func() {
							if ui.Button(c, "打开数据目录").Clicked() {
								if err := s.callbacks.OpenDataDir(); err != nil {
									s.fail(c, err)
								}
							}
							if ui.Button(c, "导出 OPML…").Clicked() {
								path, err := s.callbacks.ExportOPML(s.window)
								if err != nil {
									s.fail(c, err)
								} else if path != "" {
									s.succeed(c, "已导出订阅列表")
								}
							}
						})
					})
				})
			})
		})

		ui.Box(c).Height(1).Shrink(0).Background(theme.Border)
		ui.Row(c).Shrink(0).Padding(12, 20).Gap(12).AlignItems(ui.Center).Children(func() {
			switch {
			case s.err != "":
				ui.Text(c, s.err).TextColor(theme.Danger).Grow(1)
			case s.notice != "" && time.Now().Before(s.noticeAt):
				ui.Text(c, s.notice).TextColor(theme.TextMuted).Grow(1)
			default:
				ui.Textf(c, "更改会自动保存").TextColor(theme.TextMuted).Grow(1)
				s.notice = ""
			}
			if ui.Button(c, "关闭").Clicked() && s.window != nil {
				s.window.Close()
			}
		})
	})

	if changed {
		s.save(c)
	}
}

// refreshStorage reloads the disk usage snapshot at most every five seconds
// so the window stays current while it is open.
func (s *settingsView) refreshStorage(c *ui.Context) {
	if s.callbacks.StorageUsage == nil {
		return
	}
	if !s.storageAt.IsZero() && time.Since(s.storageAt) < 5*time.Second {
		return
	}
	s.storage = s.callbacks.StorageUsage()
	s.storageAt = time.Now()
	c.After(5*time.Second + 10*time.Millisecond)
}

// save persists the current draft and refreshes the form with normalized
// values.
func (s *settingsView) save(c *ui.Context) {
	stored, err := s.callbacks.Save(s.prefs)
	if err != nil {
		s.fail(c, err)
		return
	}
	s.prefs = stored
	s.err = ""
	s.succeed(c, "已保存")
}

// succeed shows a short confirmation in the footer.
func (s *settingsView) succeed(c *ui.Context, message string) {
	s.err = ""
	s.notice = message
	s.noticeAt = time.Now().Add(2 * time.Second)
	c.After(2*time.Second + 10*time.Millisecond)
	c.Invalidate()
}

// fail shows an inline error and keeps the current draft editable.
func (s *settingsView) fail(c *ui.Context, err error) {
	s.notice = ""
	s.err = fmt.Sprintf("操作失败：%v", err)
	c.Invalidate()
}

// dbSizeDescription renders the database size line of the data fieldset.
func dbSizeDescription(info service.StorageInfo) string {
	if info.DBBytes <= 0 {
		return "数据库大小：—"
	}
	return "数据库大小：" + format.Bytes(info.DBBytes)
}

// cacheUsageDescription renders the image cache usage line under the cache
// limit field.
func cacheUsageDescription(info service.StorageInfo) string {
	if info.ImageCacheLimitBytes <= 0 {
		return ""
	}
	usage := "—"
	if info.ImageCacheBytes > 0 {
		usage = format.Bytes(info.ImageCacheBytes)
	}
	return "当前缓存占用：" + usage + " / 上限 " + format.Bytes(info.ImageCacheLimitBytes)
}

// indexOf returns the index of value, or zero when it is absent.
func indexOf(values []string, value string) int {
	for i, candidate := range values {
		if candidate == value {
			return i
		}
	}
	return 0
}
