# AGENTS.md

Ressu 是一个免费的桌面 RSS 阅读器。Go 进程负责抓取、解析、持久化和搜索订阅数据，再通过
MyGO 绑定把数据交给内嵌静态阅读页；订阅源侧栏、文章列表、文章头部和底部状态栏由
MyGO 原生 UI 直接渲染，页面只负责阅读区正文和帮助对话框。面向用户的中文介绍在 `README.md`。

本文件是仓库级开发说明。开始改动前先读它；子目录里的 `AGENTS.md` 若存在，优先遵守更具体
的规则。

## 快速约束

- 包管理器固定为 Bun，锁文件是 `bun.lock`。不要引入 `package-lock.json` 或 `pnpm-lock.yaml`。
- Go 代码、标识符、代码注释、文档注释和提交信息用英文；界面字符串、toast、README 和
  `AGENTS.md` 用简体中文。
- 提交遵循 Conventional Commits，例如 `feat: ...`、`fix(ui): ...`、`docs: ...`。破坏性
  改动加 `!`。
- WebView 前端不得直接访问网络（`fetch`、`XMLHttpRequest`、`WebSocket`、`EventSource`）。
  新数据入口必须走 Go 绑定。
- 不要手改生成物：`build/`、`.mygo/`。
- 只改与任务相关的代码。不要顺手修复无关失败，也不要在任务中新增 ESLint/Prettier 配置。
- `gofmt` 必须干净。

## 技术栈

| 层     | 选择                                                                 |
| ------ | -------------------------------------------------------------------- |
| 外壳   | MyGO 0.8.3（`github.com/egoist/mygo`），Go，无 cgo，使用系统 WebView |
| 后端   | Go 1.27.3，`mmcdole/gofeed` 1.5，bbolt 1.5                           |
| 前端   | 无框架静态 HTML/CSS/JS，内嵌系统 WebView                             |
| 构建   | MyGO 内嵌 `webview/static`                                           |
| 状态   | 静态阅读页内的局部状态；没有路由、Redux 或服务端渲染                  |
| 工具链 | Bun、Go test、Vet                                                    |
| 存储   | 用户数据目录下的 bbolt 单文件数据库 `ressu.db`                       |
| 平台   | 以 Windows 为主，同时支持 Linux 与 macOS                             |
| 许可证 | 0BSD                                                                 |

仓库没有 `Cargo.toml` 或 `pyproject.toml`。两个清单是 `package.json` 和 `go.mod`
（`module github.com/sdttttt/ressu`）。

## 目录结构

```text
main.go                    应用入口：打开数据库、组装 action/webendpoint、创建窗口、启动刷新
menu_darwin.go             macOS 系统菜单；设置直接打开原生窗口，其余动作广播 ui.command
menu_other.go              Windows/Linux：SetMenu(nil)，移除原生菜单栏
internal/
  rss/                     HTTP 抓取与 RSS/Atom/JSON Feed 解析
    fetch.go               ValidateURL、Fetch、ETag/If-Modified-Since、超时与体积上限
    parse.go               Parse、Meta/Item、相对链接解析、plainText/truncate
    testdata/              RSS 2.0、Atom、JSON Feed、坏 XML 夹具
  icon/                    订阅源图标发现、下载校验与体积限制
  sanitize/                基于 bluemonday 的订阅源正文清洗
  store/                   bbolt 持久化、订阅与偏好缓存、未读计数
    store.go, model.go, icons.go, migrate.go
  media/                   正文图片安全下载、启动时清空重用的磁盘缓存和本地协议
  service/                 业务逻辑、事件与后台行为；不依赖 mygo 或 gui
    feeds.go, posts.go, settings.go, snapshot.go, events.go, foreground.go
  action/                  应用动作层：组合 service、UI 副作用与状态广播
    action.go              AppActions、UI 回调、OPML/图片/快照/偏好/外链动作
  webendpoint/             页面绑定适配层：只做 MyGO 调用与调用方窗口提取
    webendpoint.go         WebEndpoint（页面名 Feeds）、Posts、Settings、Status 绑定对象
  event/                   统一广播 DTO、eid、串行发布器和原生状态镜像
  status/                  原生状态栏与页面共享的服务状态模型
  buildinfo/               从 Go 构建元数据读取构建日期
  gui/                     MyGO 原生界面：订阅源栏、文章列表、状态栏、通知和原生窗口
    app_view.go            主窗原生布局：侧栏、文章列表、嵌入阅读页与状态栏
    feed_sidebar.go        原生订阅源栏：搜索、选择、刷新、未读和右键菜单
    post_list.go           原生文章列表：筛选、搜索、排序、分页和键盘选择
    post_header.go         原生文章头部：标题、元信息、星标、已读和原文按钮
    post_header_icons.go   文章头部使用的内嵌 SVG 图标
    format.go              原生文章列表使用的相对时间与字数格式化
    about_window.go, status_bar.go, image_viewer.go, settings_window.go, native.go, notifications.go
    assets/                原生窗口内嵌图片资源
webview/static/            无构建前端源码，由 MyGO 开发时直供并在打包时内嵌
  index.html               阅读页入口，lang="zh-CN"
  reader.css               正文排版、主题 token、空态和帮助对话框样式
  reader.js                绑定调用、事件应用、快捷键、链接和图片拦截
resources/icon.png         MyGO 打包图标
mygo.json                  MyGO CLI 配置
.github/workflows/ci.yml   测试、打包与 GitHub Release
```

`build/`、`.mygo/`、`coverage/` 和 `node_modules/`
都是生成物或本地依赖，已被 git 忽略。

## 构建与测试

首次安装：

```sh
bun install
```

| 命令                 | 作用                                                     |
| -------------------- | -------------------------------------------------------- |
| `bun run dev`        | `mygo dev`：构建并运行 `.mygo/dev` 下的开发版             |
| `go test ./...`      | Go 单测与 httptest 集成测试                              |
| `go vet ./...`       | Go 静态检查                                              |
| `bun run build`      | `mygo build`：内嵌静态前端、编译并打包到 `build/`        |
| `bunx mygo vet`      | MyGO 的 Go vet 与原生 UI 生命周期检查；未接入 CI         |
| `bunx mygo doctor`   | 检查开发环境                                             |

注意：

- 全新克隆后先 `bun install`，供 MyGO CLI 使用。
- 运行中的 `mygo dev` 可能与 `mygo generate` 冲突并导致生成超时；重新生成前先退出开发版。
- `mygo dev` 监听 Go 改动并重建应用；开发终端里 `r` 重建、
  `c` 清屏、`q` 退出。
- `bun run build` 不传 `-platform` 时使用 MyGO 的默认目标；发布必须显式传入目标平台。CI
  分别构建 `windows/amd64`、`linux/amd64`、`darwin/universal`。
- 环境要求：Go 1.27.3+、Bun、系统 WebView。Windows 10 需要 WebView2 Runtime；
  Linux 需要 WebKitGTK 4.1 与 AppIndicator 相关包。

## 启动与后台刷新

`main.go` 的启动顺序：

1. `mygo.App.SetName("Ressu")`，解析 `mygo.PathUserData`，执行 `store.Open(dir)`。
2. 创建 `action.AppActions`（组合 service、Hub 和原生 UI 适配器），再绑定
   `webendpoint` 中的 `Feeds`、`Posts`、`Settings`、`Status` 四个适配对象。
3. 创建 `ForegroundRefresher` 并挂到应用激活/失活和窗口焦点事件。
4. `WhenReady` 中应用保存的主题、安装菜单、创建主窗口；只有
   `RefreshOnStart` 为真时才在窗口显示后自动刷新最近 5 分钟未抓取的订阅源。
5. `app.Run()` 阻塞；启动阶段错误走 `log.Fatal`。

主窗口为 1200x800，最小 900x600，`StateKey: "main"`，由 MyGO 写入
`window-state.json`。窗口先以 `Hidden` 创建，在 `OnReadyToShow` 中显示；5 秒内页面仍未
就绪也会兜底显示，避免空白窗口。

`internal/service/foreground.go` 记录应用是否真正离开前台。应用失活后再次激活时，会异步
执行一次 `RefreshStale`；窗口失焦/聚焦只在应用本身失活/激活时作为兜底信号。
`RefreshStale` 只抓取 `LastFetchedAt` 为空或早于 5 分钟的订阅源，并复用全量刷新的单飞和
并发逻辑；已有刷新时返回 `ErrRefreshing`。手动全量或单源刷新不受节流限制。

## 分层与 MyGO 绑定

数据自上而下走 `webview -> webendpoint -> action -> service`，副作用自下而上由 `action`
回调给原生层：

- `internal/service` 只有业务逻辑，不 import `mygo`，也不 import `internal/gui`。
- `internal/action` 是所有页面和原生界面能力的唯一应用入口，持有 service、Hub 和 `UI`
  回调接口，负责文件对话框、剪贴板、外链、截图、原生主题与状态广播。需要暴露给界面层
  的能力必须先定义成 `AppActions` 方法。
- `internal/gui` 同时承载原生界面和 `NativeUI` 副作用实现；MyGO 对话框、shell、剪贴板、
  主题和窗口截图封装在这里。
- `internal/webendpoint` 只持有 `*action.AppActions`，负责 MyGO 绑定适配：用
  `mygo.CallerWindow(ctx)`、`mygo.CallerPage(ctx)` 取出调用方窗口/页面后转交 action。
  它不得直接持有或调用 `service`，不得复制业务逻辑，也不得自行拼装 UI 副作用。
- `internal/gui` 只依赖 `gui.SettingsCallbacks` 和 `store`，不 import `service`。
- 绑定对象上的导出方法都能被页面调用。首参 `context.Context` 由 MyGO 注入；静态页通过
  `globalThis.mygo.call` 调用时省略上下文，例如 `Feeds.Add(ctx, url)` 对应
  `runtime.call("Feeds.Add", url)`。
- 参数和返回值必须可 JSON 序列化。可空时间用 `*time.Time`，页面收到 `string | null`。
- Go 只声明一个广播通道 `app:event`，载荷是 `internal/event.Event`。所有 EID 共用同一
  结构；`eid` 决定读取 `sync`、`feeds`、`posts`、`appearance`、`prefs`、`selection`、
  `postSelection`、`post` 或 `command` 中的哪一段，未使用字段为 `null` 或空字符串。
- 事件由 `internal/event.Hub` 串行发布：先分配单调递增的 `revision`、更新时间，再更新
  `status.Model`，最后广播。不要绕过 Hub 直接调用 `AppEvent.Broadcast`。
- 主要 EID 有 `state.snapshot`、`sync.started`、`sync.progress`、`sync.finished`、
  `feeds.changed`、`posts.changed`、`appearance.changed`、`prefs.changed`、
  `selection.changed`、`post.selected`、`post.updated` 和 `ui.command`。
  `state.snapshot` 在 `Status.PageReady` 时补发完整状态。
- `feeds.changed` 携带完整订阅列表和总未读数；`posts.changed` 只携带 feed id 和新文章数。
- `post.selected` 携带选择键和 `store.PostSummary`，同步原生文章列表、原生文章头部和页面
  阅读区；`post.updated` 携带 `store.PostSummary`，同步单篇文章的已读和星标状态。
- `prefs.changed` 携带完整 `store.Prefs`，原生设置窗口保存后广播，页面据此同步偏好并在
  `postsPerFeed` 变化时重载文章。
- macOS 原生菜单的“设置”直接调用 `action.OpenSettings(window)` 打开原生设置窗口；
  `add-feed`、`refresh` 仍广播命令字符串给页面，页面在 `reader.js` 的单一 dispatcher 中
  执行动作。Windows/Linux 没有原生菜单栏。
- 任何会改变订阅或未读状态的服务方法都应调用 Hub 的 `FeedsChanged`；多窗口依赖该事件
  保持一致。

## 原生订阅源栏与文章列表

- 主窗口由 `internal/gui/app_view.go` 组合两层原生 `ui.SplitPercent`：外层是订阅源栏和
  阅读区，内层是文章列表和 WebView 正文。`Prefs.PaneSizes[0]` 是侧栏占整窗宽度的百分比，
  `Prefs.PaneSizes[1]` 是文章列表占阅读区宽度的百分比；拖拽停止 500ms 后分别由
  `action.SetSidebarPercent` 和 `action.SetListPercent` 保存。
- `FeedSidebar`、`PostList` 和 `PostHeader` 都只持有 `status.Model` 快照与 `AppActions`
  提供的回调。原生文章列表自己维护筛选、搜索、排序和分页加载，每页 100 篇；选择文章时携带
  完整摘要发布 `post.selected`，已读和星标变化由服务层发布 `post.updated`。原生头部渲染
  当前摘要，并提供星标、已读/未读和原文操作。
- 原生界面主题从 `ui.LightTheme()`/`ui.DarkTheme()` 重新构建，不能复制嵌套
  `ui.Themed` 里的 `c.Theme()` 后再缩放字号，否则会重复应用系统/应用字号。侧栏和文章
  列表都以 13px 为基准字号，文章列表使用与旧 WebView 完全一致的 OKLCH 色彩 token。
- 文章列表通过 `OnWindowShortcut` 处理 `j`/`k`/`n`/`p`/`Enter`/`s`；文本输入获得焦点时
  跳过。静态正文只处理 `Space`，滚到底部后通过 `Posts.SelectNeighbour` 交给原生列表继续
  选择。

## 原生设置窗口

偏好设置不再由页面里的对话框承载，而是 `internal/gui/settings_window.go` 的原生窗口：

- 入口是静态页的 `Settings.OpenSettings` 命令或 macOS 菜单直接调用
  `action.OpenSettings(window)`；`action` 把它转成 `UI.OpenSettings(parent, callbacks)`，
  由 `internal/gui/native.go` 打开原生窗口。
- 窗口是主窗口的子窗口（`Parent`），`StateKey: "settings"`，`UseContentSize` 580x720，
  最小 480x560，内容可滚动；同时只允许存在一个实例。
- 控件通过 `gui.SettingsCallbacks`（`Load`/`Save`/`DataDir`/`ExportOPML`/`OpenDataDir`）
  读写数据，`gui` 因此不依赖 `service`。
- 每次改动立即保存并广播 `prefs.changed`。保存时用 `action.UpdateSettingsDraft` 保留当前的
  `PaneSizes` 和 `SidebarCollapsed`，避免覆盖页面侧的布局改动。
- 原生控件是即时模式：构造器写入指针，用 `Element.Changed()` 判断变化，用
  `c.Invalidate()` 触发重绘；不要把 `ui.Context` 或控件存进状态或后台 goroutine。

## 原生关于窗口

- 状态栏版本号是普通文本，点击由 `Element.OnClick` 绑定，不使用按钮控件；`main.go`
  组装元数据并在点击时补入当前主题和字号。
- 点击后打开 `internal/gui/about_window.go` 的单例子窗口，显示版本、嵌入头像、开发者、
  构建日期、邮箱和说明文字；邮箱使用 `mailto:` 链接，窗口只通过系统标题栏关闭。
- 构建日期由 `internal/buildinfo.Date` 从 Go 的 `vcs.time` 读取；缺少 VCS 信息时回退到
  可执行文件的修改时间。

## 原生通知

- 页面不使用网页 toast 组件。静态页通过 `Status.Notify(level, message)` 调用，再经
  `webendpoint`、`AppActions.Notify` 和 `gui.NativeUI.ShowNotification` 进入原生队列。
- 通知绑定在非 UI goroutine 上执行，不能直接操作 `ui.Context`。`internal/gui/notifications.go`
  先加锁入队，并调用主窗口的 `Invalidate()`；`AppView.View` 在下一帧取出队列后调用
  `ui.ToastViewportBase` 渲染。
- 通知队列最多保留 32 条，类型只接受 `error`、`warning`、`info`、`success`；通知卡片固定在
  状态栏上方的右下角，使用当前主题和字号，MyGO 最多同时显示 3 条。提示文案保持简体中文。

## 订阅刷新流程

- `Feeds.Add` 先校验绝对 `http`/`https` URL、检查重复，再立即抓取一次并保存频道元数据与
  首批文章；成功后广播 `feeds.changed` 和 `posts.changed`。
- `Feeds.RefreshAll` 供手动全量刷新使用，用 `maxConcurrent = 4` 的信号量并发抓取，每个
  订阅完成后广播一次 `sync.progress`，批次开始和结束分别广播 `sync.started` 与
  `sync.finished`。
- `Feeds.RefreshStale` 供启动和回前台自动刷新使用：按 `Feed.LastFetchedAt` 过滤最近 5 分钟
  已经抓取过的订阅源，再交给同一并发刷新逻辑；没有到期订阅源时直接返回。
- `refreshOne` 把 ETag 与 Last-Modified 交给 `rss.Fetch`。304 映射为
  `rss.ErrNotModified`：保留文章，只清除错误并更新 `lastFetchedAt`。
- 抓取失败会把错误写入 `Feed.LastError`，保留订阅和已存文章，并广播
  `feeds.changed`；原生状态栏展示错误摘要，页面通过原生通知补充提示。
- 成功抓取后调用 `store.MergePosts`，再刷新频道元数据、校验器和 `lastFetchedAt`。合并完成
  后重新读取 Feed，因为未读数已经在事务中重算。
- `Posts.List` 在服务端完成未读/星标筛选、搜索和分页。返回 `{posts,total,unread}`。
  界面每页 100 条；服务默认 100、上限 500。搜索已实现，但当前界面始终传空字符串。
- 未读数由文章推导并保存在 `Feed.Unread`：合并时重算，单条已读状态改变时增减，全部已读时清零。

## 持久化

数据库是用户数据目录下的 `ressu.db`（bbolt，单文件，文件模式 0600）：

| 桶         | 键                | 值                                                      |
| ---------- | ----------------- | ------------------------------------------------------- |
| `feeds`    | 8 字节大端订阅 id | 整个 `Feed` 的 JSON：元数据、ETag/Last-Modified、未读数 |
| `posts`    | 订阅 id（嵌套桶） | 每个 GUID 一条 `Post` 的 JSON                           |
| `settings` | `prefs`           | 整个 `Prefs` 的 JSON                                    |
| `icons`    | 订阅 id（嵌套桶） | `data` 键存原始图片字节；`meta` 键存缓存元数据 JSON     |

`window-state.json` 由 MyGO 依据 `StateKey` 写在同一个用户数据目录，不是本仓库代码。

- `Store` 在内存缓存订阅和偏好，文章每次从数据库读取；一把互斥锁串行化写操作。返回的
  slice 会克隆，调用方不能改到内部状态。
- bbolt 的文件锁意味着同一数据目录只能有一个进程打开数据库；第二个实例会超时报错。
- `settings/schemaVersion` 记录数据库版本。修改持久化结构时必须兼容旧记录，或在
  `store/migrate.go` 加入迁移。
- `MergePosts` 按 GUID 合并：已有文章保留 `read`、`starred`、`fetchedAt`，其余字段刷新；
  按 `publishedAt` 倒序，裁剪到 `Prefs.PostsPerFeed`，并重算未读。它在一个事务里重写该
  订阅的整个文章桶。
- `UpdatePost` 和 `MarkAllRead` 只写受影响的数据，并同步更新内存中的未读数。
- `AddFeed` 使用“当前最大 id + 1”；删除最高 id 后该 id 会被复用。
- `Prefs.Normalize` 会替换越界值。未知主题回落 `system`，未知字号回落 `medium`，
  `openLinks` 不是 `window` 就变成 `browser`，`remoteResources` 不是 `webview` 就回落
  `goProxy`，`postsPerFeed` 限制在 50–2000，
  面板宽度不合规则恢复 `[18,30]`。
- 页面只消费 `prefs.changed` 里的 `theme`、`fontSize` 和 `markReadOnOpen`；偏好默认值只以
  `store.DefaultPrefs` 为准。
- `Post.Content` 是入库前清洗后的 HTML，`Post.Summary` 是截到 280 个字符的纯文本。
  页面渲染时不再重复清洗正文。

## 订阅源抓取与解析

- `rss.ValidateURL` 只接受带 host 的绝对 `http`/`https` URL；`file:`、`javascript:` 和相对
  地址都返回 `ErrInvalidURL`。
- `rss.Fetch` 发送 `User-Agent: Ressu/0.1 (+https://github.com/sdttttt/ressu)`、偏向订阅源的
  `Accept`，以及可用的 `If-None-Match`/`If-Modified-Since`。
- 请求和读取 body 共享 15 秒超时；响应最多读取 16 MiB；非 200/304 状态直接报错。
- `rss.Parse` 使用 gofeed 读取 RSS 0.9x/1.0/2.0、Atom 和 JSON Feed。GUID 缺失时先退回
  链接，再退回“标题 + 发布时间”；标题也缺失时使用 GUID。
- 频道和文章里的相对链接会按订阅源地址解析为绝对地址。摘要会去掉 HTML、折叠空白并截断。
- 正文在 `internal/sanitize` 基于 bluemonday 清洗，随后存入数据库。链接和媒体地址会按
  文章地址补全为绝对地址。
- 损坏 XML 或不是订阅源的页面返回错误，服务层把错误展示给用户。
- 订阅源图标由 `internal/icon` 在 Go 端解析：优先 feed 内置图标，其次站点 favicon
  链接，最后 `/favicon.ico`。只允许位图、http/https、256KB 以内，并通过 bbolt 缓存。
- `Feeds.Icon` 使用 7 天 TTL 和按订阅源 singleflight；显式图标地址变化时缓存失效。

## 前端结构与状态

- `webview/static` 不经过打包器。`reader.js` 是唯一入口脚本，局部状态只保存当前文章、
  订阅列表、事件 revision 和影响阅读的偏好。
- 所有页面动作走 `globalThis.mygo.call`；只有错误提示转发给 `Status.Notify`。页面不能引入
  网络 API 或第三方运行时。
- `runtime.on("app:event")` 在启动时订阅一次；唯一 dispatcher 按 `eid` 路由，并丢弃
  `revision` 不递增的事件。
- 文章请求用递增 requestId 防止快速切换时旧响应覆盖新文章。`post.selected` 获取正文后，
  若偏好要求则自动标记已读。
- 订阅源栏、文章列表、文章头部和偏好设置都是原生 UI；页面只渲染正文、空态、加载态和帮助
  对话框。
- 页面快捷键只有 `Space`、`m`、`v`、`h`、`l` 和 `?`；原生列表已处理 `j`、`k`、`n`、`p`、
  `Enter` 和 `s`，不要在页面重复注册。
- 主题有两层：`action.UpdatePreferences` 调用 `UI.ApplyTheme` 设置原生外观；前端
  `reader.js` 在 `<html>` 上切换 `dark` 和 `theme-sepia` 类。字号通过
  `--app-text-scale` 缩放根字体。
- `sepia` 运行时类名必须是 `theme-sepia`，避免与 WebView 内置滤镜语义混淆。
- 新增主题要同步 `store` 的主题常量与 `Normalize`、`internal/gui/native.go` 的 `ApplyTheme`、
  `reader.js`（`isDark`/`applyAppearanceFromPrefs`）和 `internal/gui/settings_window.go`
  的选项。
- 需要展示日期时使用 `zh-CN` Intl；需要测试确定性时把 `now` 或时间值作为参数传入。

## 文章快照

快照入口当前已停用；`internal/action`、`internal/webendpoint`、`internal/service/snapshot.go`
和相关测试保留，便于后续恢复。恢复时不能用当前窗口截图代替长文快照：仍需重建固定宽度、
固定浅色的离屏快照卡片，等待图片解码，按滚动分片调用 `Posts.CaptureSnapshot`，再由
`Posts.StitchSnapshot` 拼接。恢复前必须重新提供 UI 入口和前端卡片版式。

## 代码约定

### Go

- 包注释写在主文件；每个导出标识符都要有以自身名字开头的文档注释。
- 哨兵错误是包级 `errors.New` 值（`store.ErrNotFound`、`store.ErrDuplicate`、
  `rss.ErrNotModified`、`rss.ErrInvalidURL`、`service.ErrRefreshing`）。需要上下文时用 `%w`
  包装，错误文本要短且可直接展示在中文通知里。
- 每个持久化或跨绑定字段都要有 camelCase `json` tag。
- 跟随标准库风格：`slices`、`min`/`max`、泛型、`for i := range n`。改过的 Go 文件运行
  `gofmt -w`。
- 平台差异用 `//go:build` 拆文件（如 `menu_darwin.go` / `menu_other.go`），不要在代码里
  判断 `runtime.GOOS`。
- `Store` 方法保持加锁；返回切片时克隆。
- Go 测试与被测代码同目录，命名为 `*_test.go`。不使用 mock 库，优先
  `t.TempDir()`、`httptest` 和真实数据库。

### 静态 WebView

- 文件和目录用 kebab-case；不引入框架、打包器、UI 库或 CSS 框架。
- 页面字符串用简体中文；JavaScript 标识符和注释用英文。
- 样式使用 `reader.css` 的语义变量和系统主题类，不要写死主题颜色。
- `Post.Content` 是 Go 已清洗的 HTML，可以直接进入正文容器；其他不可信文本用
  `textContent` 或安全节点写入。
- 新增交互时先确认原生 UI 是否已经承载该能力；只有阅读区能力才加入静态页。

## 常见改动路径

- 新增后端调用：先在 `internal/service` 写业务逻辑（不碰 MyGO），再在
  `internal/action` 增加 `AppActions` 方法；必要时把新的系统交互封装成 `internal/gui`
  方法并实现 `action.UI`。只有 `AppActions` 负责事件广播和 UI 副作用，最后在
  `internal/webendpoint` 的对应绑定对象上增加只做参数/调用方窗口适配并转发到 action 的
  导出方法；在 `reader.js` 的对应交互中调用，处理错误和局部状态。
- 新增事件：在 `internal/event/event.go` 加 EID 和字段，补 `Validate` 与 Hub 发布方法，
  在状态变化处调用；在 `reader.js` 的 dispatcher 加分支。不要新增第二个
  `mygo.NewEvent` 通道。
- 新增偏好：修改 `store.Prefs`、`DefaultPrefs`、`Prefs.Normalize` 和前端
  相关消费点；在 `internal/gui/settings_window.go` 加原生控件；若影响阅读则更新
  `reader.js` 和应用逻辑。
- 新增快捷键：先确认属于原生列表还是静态正文；原生快捷键接在 `post_list.go`，正文快捷键
  接在 `reader.js`，帮助对话框保持中文说明完整。
- 新增主题：同步 Go 常量/校验、原生主题、CSS 变量、`reader.js`、原生设置窗口和安全测试。
- 新增界面文案：中文、语义 token、`~` 导入、默认导出；除非任务明确要求，不引入 i18n 层。

## 测试

- Go：`go test ./...`。`rss` 是表驱动测试，夹具在 `internal/rss/testdata/`；`store` 覆盖
  数据库创建、重复拒绝、合并语义、裁剪、全部已读、删除、重开和图标缓存；`service` 用
  httptest 跑添加、刷新、304、自动刷新节流、新增、标记已读、筛选、分页、搜索和图标缓存；
  `event` 覆盖 EID 字段校验、revision 顺序、并发发布和完整状态快照；
  `icon` 覆盖图标候选、下载校验和体积上限；`foreground` 覆盖回前台刷新和窗口焦点兜底；
  `snapshot` 覆盖请求校验、拼接、错误分片、文件名和成品 PNG 校验；`action` 用假 `UI`
  覆盖外链校验、偏好草稿保存和通知转发；`gui` 用 `ui.NewTester` 渲染原生订阅源栏、文章
  列表、头部、通知、设置和关于窗口，并验证状态栏版本点击。
- 静态页没有独立测试框架；语法可用 `node --check webview/static/reader.js` 检查，运行行为
  由打包后的应用手动验证。
- 没有 E2E、浏览器自动化或覆盖率阈值。

提交前至少运行与改动相关的测试；涉及跨层行为时运行：

```sh
go test ./...
go vet ./...
bun run build
```

Go 集成测试会绑定本地 httptest 端口。在受限沙箱中需要允许本地监听，否则会出现
`bind: operation not permitted`。

## CI 与发布

`.github/workflows/ci.yml`：

- `test` 在推送 `master`/`develop`、推送 `v*` tag 和向这两个分支提 PR 时运行，矩阵为
  Ubuntu + Windows。步骤依次是 `bun install --frozen-lockfile`、`go test ./...`、
  `node --check webview/static/reader.js` 和 `bun run build`。
- `release-version` 只对 `v*` tag 运行，校验 tag 符合 `vYY.M.N`，且与 `package.json`、
  `mygo.json` 的版本号一致。`M` 是 1–12 且不补零，`N` 从 1 开始。
- `package` 依赖测试和版本校验，在 Windows、Linux、macOS 上分别构建
  `windows/amd64`、`linux/amd64`、`darwin/universal`，上传 `build/`。
- `release` 汇总 DMG、Windows setup EXE、DEB、tar.gz，生成 `SHA256SUMS`，再发布 GitHub
  Release。tag 中含 `-` 时标记为预发布。
- 发版时同时修改 `package.json` 与 `mygo.json` 的版本号，并打对应 tag。
- 目前没有自动更新配置；MyGO 更新签名密钥应通过环境变量提供，不要提交到仓库。

## 安全

- 订阅源内容不可信。正文必须在 Go 端经过 `internal/sanitize` 的 bluemonday 策略清洗后
  入库，并禁用 `script`、`style`、`iframe`、`object`、`embed`、`form`、`input`、
  `button` 以及 `style`、`srcset`、`onerror`、`onload`。它还会把链接改为绝对地址并加
  `target="_blank" rel="noopener noreferrer"`，图片改为懒加载。
- `reader.js` 的正文容器和未来快照卡片里的 `innerHTML` 只允许接收数据库中已清洗的正文；
  不要把抓取原文直接传给前端。嵌入 WebView 的快照必须通过 `Page.WebView().CapturePage()`
  捕获。
- 打开外链前必须经过 `externalHref()`，只允许 `http:`/`https:`。`action.OpenLink`
  再校验一次绝对 `http`/`https` 地址，并按偏好调用系统浏览器或创建应用内窗口；不要把原始
  href 直接交给 shell 或窗口。
- 订阅地址在 Go 侧由 `rss.ValidateURL` 再校验一次；不要只依赖前端校验。
- 数据是本地未加密数据库，包含订阅、完整正文、已读/星标和偏好；原生设置窗口可打开数据目录。
- 没有账号、同步、遥测或统计。WebView 不直接加载订阅源图标；图标由 Go 缓存后以本地 data
  URL 渲染。正文远端图片默认由 Go 在 `Posts.Get` 中改写为 `ressu-media://`，通过本地缓存协议
  提供；用户可在设置中选择 Webview 直接加载。
- webview 端禁止直接网络请求；代码评审必须确认没有 `fetch`、`XMLHttpRequest`、
  `WebSocket` 和 `EventSource`。需要远程数据时新增 Go 绑定。
- `mygo build` 从 `mygo://localhost` 提供内嵌页面；页面没有 CSP meta 标签，因此 XSS 防护
  依赖内容清洗和 URL 校验，而不是 CSP。
- 快照保存路径来自原生保存对话框；文件名由文章标题清理生成，只写入用户选中的路径。
  不要把快照上传到网络或写入订阅源可控的路径。
- 不要提交凭据、用户数据或生成物。签名/更新密钥放在环境变量中。
