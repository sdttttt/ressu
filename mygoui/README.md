# Ressu 原生界面（mygoui）

MyGO 原生界面的启动装配、组合根与副作用实现，独立为 `github.com/sdttttt/ressu/mygoui`
module，只依赖 mygo 与 `ressu/core`。根包的 `run.go` 负责启动装配（`Run`：打开数据库、
组合 service 与 `core/app.App`、注册 `webendpoint` 绑定、创建主窗），`menu_*.go` 实现
macOS 菜单；根包还做主窗布局、状态栏和 `NativeUI`；侧栏、文章列表、文章头部的回调组装
集中在根包的 `callbacks.go`；可复用的主题、格式化、图标、控件、对话框和子窗口放在基础包，
订阅源栏、文章列表、文章头部和各类窗口放在功能包。组件只持有 `status.Model` 快照与
`core/app.App` 提供的回调，不得直接访问数据库或实现业务规则。

原生界面是即时模式 UI：构造器写入指针，用 `Element.Changed()` 判断变化，用
`c.Invalidate()` 触发重绘；不要把 `ui.Context` 或控件存进状态或后台 goroutine。

## 主窗布局

- `app_view.go` 组合两层原生 `ui.SplitPercent`：外层是订阅源栏和阅读区，内层是文章列表和
  WebView 正文。`Prefs.PaneSizes[0]` 是侧栏占整窗宽度的百分比，`Prefs.PaneSizes[1]` 是
  文章列表占阅读区宽度的百分比；拖拽停止 500ms 后分别由 `app.SetSidebarPercent` 和
  `app.SetListPercent` 保存。
- 原生界面主题从 `ui.LightTheme()`/`ui.DarkTheme()` 重新构建，不能复制嵌套
  `ui.Themed` 里的 `c.Theme()` 后再缩放字号，否则会重复应用系统/应用字号。侧栏和文章
  列表都以 13px 为基准字号，文章列表使用与旧 WebView 一致的 OKLCH 色彩 token。

## 订阅源栏（sidebar）

- 搜索、选择、刷新、未读徽标和右键菜单（打开/编辑/刷新/标记全部已读/取消订阅）。行高 34、
  间距 2。
- **选中高亮是一个独立的“胶囊”元素**，不是行自绘背景：绝对定位在行容器内、
  `ZIndex(-1)` 垫在透明的行之下、`IgnoreInput(true)` 穿透指针，Y 坐标用
  `AnimateTo("y", …, selectionSlide)` 弹簧插值，快速连续切换（h/l）时弹簧保留速度连续
  滑动；`Snapshot.ReduceMotion`（设置里的“减少动效”）时用 `Jump` 直接落位。行背景
  必须保持透明，否则滑动过程会被盖住。
- 折叠态由 `DrawOverlay` 渲染浮动面板，`AnimatePose` 控制滑入滑出；减少动效时用 1ms
  定时动画代替 Snappy 弹簧直接落位。
- 快捷键（`h`/`l` 切源、`?` 帮助）由主窗口统一处理，见下文分工。

## 文章列表（postlist）

- 自己维护筛选（未读/星标）、搜索、排序（发布时间或抓取时间，`Prefs.PostSort`）和分页
  加载，每页 100 篇。
- 通过 `OnWindowShortcut` 处理 `j`/`k`/`n`/`p`/`Enter`/`s`；选择文章时携带完整摘要发布
  `post.selected`，已读和星标变化由服务层发布 `post.updated`。

## 快捷键分工

| 按键                       | 处理方                       |
| -------------------------- | ---------------------------- |
| `j` `k` `n` `p` `Enter` `s` | 文章列表（postlist）        |
| `h` `l` `?`（Shift+/）     | 主窗口（app_view）          |
| `Space` `m` `v`            | 阅读页正文（reader.js）     |

文本输入获得焦点时全部跳过。WebView 持有焦点时，`reader.js` 只拦截 `h`、`l` 和 `?`，
经 `Status` 绑定转交原生 GUI 并 `preventDefault()`，避免物理按键丢失。

## 原生设置窗口（settings）

- 入口是原生订阅源栏的菜单或 macOS 菜单直接调用 `core/app` 的 `OpenSettings(window)`；窗口是
  主窗口的子窗口，`StateKey: "settings"`，`UseContentSize` 580x720，最小 480x560，
  单例。
- 控件通过 `settings.Callbacks`（`Load`/`Save`/`DataDir`/`ExportOPML`/`OpenDataDir`）
  读写数据；每次改动立即保存并广播 `prefs.changed`。保存时用
  `app.UpdateSettingsDraft` 保留当前的 `PaneSizes` 和 `SidebarCollapsed`。

## 关于窗口（about）

状态栏版本号是普通文本，点击由 `Element.OnClick` 绑定（不用按钮控件）；点击后打开
`about` 单例子窗口，显示版本、嵌入头像、开发者、构建日期和 `mailto:` 邮箱。构建日期由
`core/buildinfo.Date` 从 Go 的 `vcs.time` 读取，缺少 VCS 信息时回退到可执行文件的
修改时间。

## 原生通知（dialog）

- 页面不使用网页 toast。静态页经 `Status.Notify` → `core/app.App.Notify` →
  `NativeUI.ShowNotification` 进入原生队列。
- 通知在非 UI goroutine 上入队并触发主窗口 `Invalidate()`；`AppView.View` 下一帧取出
  队列调用 `ui.ToastViewportBase` 渲染。队列最多 32 条，类型只接受
  `error`/`warning`/`info`/`success`，固定在状态栏上方右下角，最多同时显示 3 条。

## 文章快照（snapshotcap + viewer）

- 入口在原生文章头部，头部回调经 `core/app` 的 `RequestSnapshot(parent)` 直接打开单例采集子
  窗口（`Parent` 为主窗口，`StateKey: "snapshot-capture"`，纯页面窗口加载
  `/snapshot.html`）。
- 采集页用 `Snapshot.Selected` 读当前选中文章、`Posts.Get` 取正文后重建固定浅色、
  760px 居中的快照卡片，等待图片解码并按滚动分片调用 `Posts.CaptureSnapshot`
  （`NativeUI.CapturePage` 对纯页面窗口回退为 `Window.CapturePage`）。
- 随后页面调用 `Posts.StitchSnapshot`，`core/service/snapshot.go` 拼接完成后以主窗口为
  `Parent` 打开原生预览窗口并关闭采集窗口。拼接按几何信息里的 `cardLeft` 从整窗截图中
  裁出居中卡片；生成中阻断交互，Esc 或错误页关闭按钮经 `Snapshot.Cancel` 关窗。
- 不能用当前窗口截图代替长文快照；嵌入 WebView 的快照必须通过
  `Page.WebView().CapturePage()` 或纯页面窗口的 `Window.CapturePage()` 捕获。
