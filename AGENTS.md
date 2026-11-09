# AGENTS.md

Ressu 是一个免费的桌面 RSS 阅读器。Go 进程负责抓取、解析、持久化和搜索订阅数据，再通过
MyGO 绑定把数据交给 SolidJS 页面；页面只渲染阅读区本地状态，底部状态栏由 MyGO 原生 UI
直接渲染。面向用户的中文介绍在 `README.md`。

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
- 不要手改生成物：`webview/src/mygo.ts`、`webview/dist/`、`build/`、`.mygo/`。
- 只改与任务相关的代码。不要顺手修复无关失败，也不要在任务中新增 ESLint/Prettier 配置。
- `gofmt` 必须干净；前端改完运行 `bun run fmt`。

## 技术栈

| 层     | 选择                                                                 |
| ------ | -------------------------------------------------------------------- |
| 外壳   | MyGO 0.8.1（`github.com/egoist/mygo`），Go，无 cgo，使用系统 WebView |
| 后端   | Go 1.27.3，`mmcdole/gofeed` 1.5，bbolt 1.5                           |
| 前端   | SolidJS 1.9，Kobalte 0.13 + corvu，Tailwind CSS 4                    |
| 构建   | Vite 8 + `vite-plugin-solid`，目标 `es2022`                          |
| 状态   | 单个 Solid store；没有路由、Redux 或服务端渲染                       |
| 工具链 | Bun、TypeScript 7、oxlint、oxfmt、Vitest 5                           |
| 存储   | 用户数据目录下的 bbolt 单文件数据库 `ressu.db`                       |
| 平台   | 以 Windows 为主，同时支持 Linux 与 macOS                             |
| 许可证 | 0BSD                                                                 |

仓库没有 `Cargo.toml` 或 `pyproject.toml`。两个清单是 `package.json` 和 `go.mod`
（`module github.com/sdttttt/ressu`）。

## 目录结构

```text
main.go                    应用入口：打开数据库、绑定服务、创建窗口、启动刷新
app_view.go                主窗原生布局：嵌入阅读页与底部状态栏
menu_darwin.go             macOS 系统菜单；菜单项向页面广播 ui:command
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
  service/                 暴露给页面的服务、事件与后台行为
    feeds.go, posts.go, settings.go, snapshot.go, events.go, foreground.go
  status/                  原生状态栏与页面共享的服务状态模型
  gui/                     MyGO 原生状态栏渲染
webview/                   前端源码和独立构建配置
  index.html               Vite 入口，lang="zh-CN"
  vite.config.ts           Solid + Tailwind 插件、~ 别名、root 指回 webview/、固定 5173
  vitest.config.ts         jsdom、setup 文件、src/**/*.test.{ts,tsx}
  tsconfig.json            strict、noEmit、jsx preserve、~/* 路径别名
  src/
    main.tsx               挂载 App，首帧前先按系统主题上色
    app.tsx                三栏布局、对话框与 toast
    index.css              Tailwind 4 入口、主题 token、.post-body、.snapshot-*
    mygo.ts                生成的类型化客户端；被 git 忽略，不要编辑
    components/            页面组件、对话框、图标按钮和内联 shadcn 风格原语
    components/ui/         Kobalte/corvu + cva 的 UI 原语
    lib/                   format、sanitize、snapshot、theme、commands、utils
    stores/
      app-state.ts         状态类型、默认值、Solid store、派生值
      app.ts               供组件使用的统一导出层
      feeds.ts             订阅源选择、添加、删除和刷新
      icons.ts             订阅源图标状态、本地 data URL 渲染和失效
      posts.ts             文章分页、筛选、打开、已读和星标
      snapshot-actions.ts  快照生成、预览、保存和复制
      settings.ts          偏好、面板宽度、外链和数据目录
      selection.ts         键盘选择、相邻订阅源和原文打开
      sync-events.ts       刷新与文章变更事件的本地状态同步
      lifecycle.ts         启动加载
      ui.ts                对话框和订阅源过滤状态
      events.ts            MyGO 事件订阅
    test/app-mock.ts       组件测试共用的 Solid store mock
  test/setup.ts            Vitest 的 jsdom 垫片与网络禁令
  public/favicon.svg
resources/icon.png         MyGO 打包图标
mygo.config.ts             MyGO CLI 配置
.oxlintrc.json             oxlint 规则、插件和前端网络禁令
.oxfmtrc.json              printWidth 100、import/Tailwind 类名排序
.github/workflows/ci.yml   测试、打包与 GitHub Release
```

`build/`、`webview/dist/`、`.mygo/`、`webview/src/mygo.ts`、`coverage/` 和 `node_modules/`
都是生成物或本地依赖，已被 git 忽略。

## 构建与测试

首次安装：

```sh
bun install
```

| 命令                 | 作用                                                     |
| -------------------- | -------------------------------------------------------- |
| `bun run dev`        | `mygo dev`：启动 Vite，构建并运行 `.mygo/dev` 下的开发版 |
| `bun run dev:web`    | 只启动 `http://localhost:5173` 的 Vite，没有 Go 数据     |
| `bun run generate`   | 根据 Go 服务和事件重新生成 `webview/src/mygo.ts`         |
| `bun run typecheck`  | `tsc -p webview/tsconfig.json --noEmit`                  |
| `bun run lint`       | oxlint                                                   |
| `bun run fmt`        | oxfmt，写入文件并排序 import/Tailwind 类名               |
| `bun run fmt:check`  | 只检查格式                                               |
| `bun run check`      | typecheck + lint + fmt:check；不跑测试                   |
| `bun run test`       | Vitest（jsdom + `@solidjs/testing-library`）             |
| `bun run test:watch` | Vitest 监听模式                                          |
| `go test ./...`      | Go 单测与 httptest 集成测试                              |
| `go vet ./...`       | Go 静态检查                                              |
| `bun run build:web`  | 类型检查后构建前端到 `webview/dist/`                     |
| `bun run build`      | `mygo build`：构建前端、编译并打包到 `build/`            |
| `bunx mygo vet`      | MyGO 的 Go vet 与原生 UI 生命周期检查；未接入 CI         |
| `bunx mygo doctor`   | 检查开发环境                                             |

注意：

- 全新克隆后要先 `bun run generate`，否则类型检查会因缺少 `webview/src/mygo.ts` 失败。
  `mygo dev` 和 `mygo build` 会自动重新生成它。
- 运行中的 `mygo dev` 可能与 `mygo generate` 冲突并导致生成超时；重新生成前先退出开发版。
- Vite 端口固定为 5173（`strictPort`），必须与 `mygo.config.ts` 的 `devUrl` 一致。
- `mygo dev` 监听 Go 改动并重建应用；前端热更新由 Vite 负责。开发终端里 `r` 重建、
  `c` 清屏、`q` 退出。
- `bun run build` 不传 `-platform` 时使用 MyGO 的默认目标；发布必须显式传入目标平台。CI
  分别构建 `windows/amd64`、`linux/amd64`、`darwin/universal`。
- 环境要求：Go 1.27.3+、Bun、系统 WebView。Windows 10 需要 WebView2 Runtime；
  Linux 需要 WebKitGTK 4.1 与 AppIndicator 相关包。

## 启动与后台刷新

`main.go` 的启动顺序：

1. `mygo.App.SetName("Ressu")`，解析 `mygo.PathUserData`，执行 `store.Open(dir)`。
2. 绑定 `Feeds`、`Posts`、`Settings` 三个服务。
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

## MyGO 绑定与事件

- 服务上的导出方法都能被页面调用。首参 `context.Context` 由 MyGO 注入，TypeScript 中省略：
  `Feeds.Add(ctx, url)` 生成 `Feeds.add(url)`。
- 参数和返回值必须可 JSON 序列化。可空时间用 `*time.Time`，对应 TypeScript
  `string | null`。
- `mygo generate` 生成接口、`Feeds`/`Posts`/`Settings` 客户端和 `events.*`。生成文件里的
  JSON tag 会转成 camelCase，并保留 Go 文档注释。
- 事件名用冒号分区，声明在 `internal/service/events.go`：
  `sync:state`、`feeds:changed`、`posts:changed`、`ui:command`。
- `feeds:changed` 携带完整订阅列表和总未读数；`posts:changed` 只携带 feed id 和新文章数。
- macOS 原生菜单只广播命令字符串（`add-feed`、`refresh`、`settings`），不直接拥有对话框。
  页面通过 `webview/src/stores/events.ts` 打开对话框或执行动作。Windows/Linux 没有原生菜单栏。
- 任何会改变订阅或未读状态的服务方法都应调用 `broadcastFeeds`：它先更新原生状态栏模型，
  再广播 `feeds:changed`；多窗口依赖该事件保持一致。

## 订阅刷新流程

- `Feeds.Add` 先校验绝对 `http`/`https` URL、检查重复，再立即抓取一次并保存频道元数据与
  首批文章；成功后广播 `feeds:changed` 和 `posts:changed`。
- `Feeds.RefreshAll` 供手动全量刷新使用，用 `maxConcurrent = 4` 的信号量并发抓取，每个
  订阅完成后广播一次 `sync:state`，最后再广播 `active: false`。
- `Feeds.RefreshStale` 供启动和回前台自动刷新使用：按 `Feed.LastFetchedAt` 过滤最近 5 分钟
  已经抓取过的订阅源，再交给同一并发刷新逻辑；没有到期订阅源时直接返回。
- `refreshOne` 把 ETag 与 Last-Modified 交给 `rss.Fetch`。304 映射为
  `rss.ErrNotModified`：保留文章，只清除错误并更新 `lastFetchedAt`。
- 抓取失败会把错误写入 `Feed.LastError`，保留订阅和已存文章，并广播
  `feeds:changed`；原生状态栏展示错误摘要，页面在 toast 中补充提示。
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
- 前端 `defaultPrefs`（`webview/src/stores/app-state.ts`）复制了 Go 默认值；新增偏好时两处
  都要改。
- `Post.Content` 是入库前清洗后的 HTML，`Post.Summary` 是截到 280 个字符的纯文本。
  前端渲染和生成快照时不再重复清洗正文。

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

- `webview/src/stores/app-state.ts` 持有唯一的 Solid store、状态类型、默认值和派生函数。
  `webview/src/stores/app.ts` 只是统一导出层；动作按功能拆在 `feeds.ts`、`posts.ts`、
  `icons.ts`、`snapshot-actions.ts`、`settings.ts`、`selection.ts`、`sync-events.ts`、
  `lifecycle.ts` 和 `ui.ts`。组件继续统一从 `~/stores/app` 导入，不要直接调用
  `setAppState`。
- 组件不直接调用 `~/mygo`；事件和快照模块除外，它们分别封装订阅和底层截图调用。
- 派生值使用普通函数，如 `currentFeed()`、`visibleFeeds()`、`totalUnread()`、
  `isSyncingFeed(id)`，不要引入 memo。
- store 动作用 `messageOf(error)` 弹中文 toast。只有调用方需要按成功与否分支时才返回布尔值
  （例如 `addFeed`）；事件处理里的“发后不管”用 `void action()`。
- 事件只在 `App` 的 `onMount` 中通过 `subscribe()` 订阅一次，并在 `onCleanup` 中退订。
- 没有路由。普通对话框由 `app.dialog` 控制；快照预览使用独立的 `app.snapshot` 状态。
- 三栏宽度保存在 `Prefs.PaneSizes`，拖拽停止 500ms 后由 `savePaneSizes` 写回。
- 文章列表采用分页追加，筛选同时支持未读和星标。列表当前选中项保存在
  `app.current.guid`。
- `commands.ts` 是快捷键的唯一注册表：每条命令声明绑定、分组、中文标签和动作。帮助对话框
  直接遍历同一数组，新增快捷键时也要同步 store 动作。macOS 的 mod 组合由原生菜单持有，
  页面用 `installShortcuts({ skipMod: isMac })` 跳过重复注册。
- 主题有两层：Go 的 `Settings.Update` 调用 `ApplyTheme` 设置原生命中外观；前端
  `lib/theme.ts` 在 `<html>` 上切换 `dark` 和 `theme-sepia` 类。字号通过
  `--app-text-scale` 缩放根字体。
- `sepia` 运行时类名必须是 `theme-sepia`。裸 `sepia` 会撞上 Tailwind 的 sepia 滤镜，
  给整页加 `filter: sepia(100%)` 并染掉所有图片。
- 新增主题要同步 `store` 的主题常量与 `Normalize`、`service.ApplyTheme`、`theme.ts`
  （`isDark`/`applyTheme`）和 `settings-dialog.tsx` 的选项。
- 日期格式化集中在 `webview/src/lib/format.ts`，使用 `zh-CN` Intl；需要确定性的函数接受
  可选 `now` 参数。

## 文章快照

快照不使用 snapdom、WASM 或第三方长图库，而是依赖系统 WebView 的原生截图能力：

1. `webview/src/lib/snapshot.ts` 在页面中挂载固定 760px 宽、固定浅色的离屏
   `.snapshot-card`。
2. 正文文本使用 `textContent`，标题和元数据也写入安全节点；Go 已清洗的 HTML 才能进入
   `innerHTML`。
3. 先等待所有图片 `decode()` 完成（失败图片会被移除），再测量 `scrollHeight`。
4. 按有重叠的滚动位置逐片调用 `Posts.CaptureSnapshot`；Go 通过 `Window.CapturePage()`
   截取当前可见页面。
5. `Posts.StitchSnapshot` 在 Go 中按第一张图的宽度推导缩放比例，裁剪卡片区域并拼接为
   完整 PNG；最大 32 片、单边 16384px。
6. store 把成品 PNG 放进 `app.snapshot`，`snapshot-dialog.tsx` 展示预览；
   `Posts.SaveSnapshot` 或 `Posts.CopySnapshot` 只消费这张成品图。
7. 保存成功后关闭预览；取消原生保存对话框或出错时保留预览。关闭预览会清空大图状态。

快照的浅色变量和版式都在 `index.css` 的 `.snapshot-card` 中。不要用屏幕截图代替卡片截图；
长文章必须按滚动分片拼接，避免只截到第一屏。

## 代码约定

### Go

- 包注释写在主文件；每个导出标识符都要有以自身名字开头的文档注释。
- 哨兵错误是包级 `errors.New` 值（`store.ErrNotFound`、`store.ErrDuplicate`、
  `rss.ErrNotModified`、`rss.ErrInvalidURL`、`service.ErrRefreshing`）。需要上下文时用 `%w`
  包装，错误文本要短且可直接展示在中文 toast 里。
- 每个持久化或跨绑定字段都要有 camelCase `json` tag。
- 跟随标准库风格：`slices`、`min`/`max`、泛型、`for i := range n`。改过的 Go 文件运行
  `gofmt -w`。
- 平台差异用 `//go:build` 拆文件（如 `menu_darwin.go` / `menu_other.go`），不要在代码里
  判断 `runtime.GOOS`。
- `Store` 方法保持加锁；返回切片时克隆。
- Go 测试与被测代码同目录，命名为 `*_test.go`。不使用 mock 库，优先
  `t.TempDir()`、`httptest` 和真实数据库。

### TypeScript / Solid

- 文件和目录用 kebab-case；组件文件用 `<name>.tsx`，组件名 PascalCase 并默认导出。仅在本文件
  使用的小组件留在同一文件，不导出。
- 从 `~/*` 导入。oxfmt 会排序 import 和 Tailwind 类名，直接运行 `bun run fmt`。
- Props 用内联对象类型；类名覆盖用 `class` + `cn(...)`；变体用 `cva`
  （参考 `components/ui/button.tsx`）。
- 只用语义 token：`bg-background`、`text-muted-foreground`、`bg-accent`、
  `border-border`、`text-warning`、`bg-error`、`rounded-md` 等。不要写死颜色。
- Tailwind 4 是 CSS-first：主题值在 `webview/src/index.css` 的 `@theme`/变量中，没有
  `tailwind.config.js`。
- 图标来自 `lucide-solid`，toast 来自 `solid-sonner`。
- 界面控件优先使用 `webview/src/components/ui/**`：输入用 `text-field`，数字用
  `number-field`，单选/多选用 `radio-group`/`toggle-group`，开关用 `switch`，提示用
  `tooltip`，计数用 `badge`，分隔用 `separator`。图标按钮用
  `~/components/icon-button`，`aria-label` 同时作为提示和可访问名。
- `webview/src/components/ui/**` 是源码内联的 shadcn 风格原语，保持 API 稳定。
- 新增测试放在 `webview/src/**/*.test.{ts,tsx}` 旁边。

## 常见改动路径

- 新增后端调用：在 `Feeds`、`Posts` 或 `Settings` 加导出方法；需要时广播变更事件；运行
  `bun run generate`；在对应的 `stores/feeds.ts`、`posts.ts` 或 `settings.ts` 动作中调用，
  处理错误和本地状态，最后从 `stores/app.ts` 导出。
- 新增事件：在 `internal/service/events.go` 用 `mygo.NewEvent[T]` 声明，在状态变化处广播，
  重新生成，并在 `stores/events.ts` 订阅。
- 新增偏好：修改 `store.Prefs`、`DefaultPrefs`、`Prefs.Normalize` 和前端
  `stores/app-state.ts` 里的 `defaultPrefs`；在 `settings-dialog.tsx` 加控件；若影响阅读则
  更新 `stores/settings.ts`、相关功能模块和应用逻辑。
- 新增快捷键：在 `lib/commands.ts` 的 `commands` 数组加一条，并确保 store 已提供对应动作。
- 新增主题：同步 Go 常量/校验、原生主题、CSS 变量、前端 `theme.ts`、设置对话框和安全测试。
- 新增界面文案：中文、语义 token、`~` 导入、默认导出；除非任务明确要求，不引入 i18n 层。

## 测试

- Go：`go test ./...`。`rss` 是表驱动测试，夹具在 `internal/rss/testdata/`；`store` 覆盖
  数据库创建、重复拒绝、合并语义、裁剪、全部已读、删除、重开和图标缓存；`service` 用
  httptest 跑添加、刷新、304、自动刷新节流、新增、标记已读、筛选、分页、搜索和图标缓存；
  `icon` 覆盖图标候选、下载校验和体积上限；`foreground` 覆盖回前台刷新和窗口焦点兜底；
  `snapshot` 覆盖请求校验、拼接、错误分片、文件名和成品 PNG 校验。
- 前端：`bun run test`。Vitest + jsdom，垫片在 `webview/test/setup.ts`。测试通过
  `vi.mock` 和固定夹具 mock `~/mygo`/`~/stores/app`；`restoreMocks: true` 已开启。
- `setup.ts` 会把网络 API stub 成抛错，测试不得访问网络，也不应依赖打包好的应用。
- 没有 E2E、浏览器自动化或覆盖率阈值。
- 即使测试全过，store/组件测试仍可能向 stderr 打印
  `You appear to have multiple instances of Solid`；这不是失败。

提交前至少运行与改动相关的测试；涉及跨层行为时运行：

```sh
bun run check
bun run test
go test ./...
go vet ./...
```

Go 集成测试会绑定本地 httptest 端口。在受限沙箱中需要允许本地监听，否则会出现
`bind: operation not permitted`。

## CI 与发布

`.github/workflows/ci.yml`：

- `test` 在推送 `master`/`develop`、推送 `v*` tag 和向这两个分支提 PR 时运行，矩阵为
  Ubuntu + Windows。步骤依次是 `bun install --frozen-lockfile`、`go test ./...`、
  `bun run generate`、`bun run check`、`bun run test`、`bun run build:web`。
- `release-version` 只对 `v*` tag 运行，校验 tag 符合 `vYY.M.N`，且与 `package.json`、
  `mygo.config.ts` 的版本号一致。`M` 是 1–12 且不补零，`N` 从 1 开始。
- `package` 依赖测试和版本校验，在 Windows、Linux、macOS 上分别构建
  `windows/amd64`、`linux/amd64`、`darwin/universal`，上传 `build/`。
- `release` 汇总 DMG、Windows setup EXE、DEB、tar.gz，生成 `SHA256SUMS`，再发布 GitHub
  Release。tag 中含 `-` 时标记为预发布。
- 发版时同时修改 `package.json` 与 `mygo.config.ts` 的版本号，并打对应 tag。
- 目前没有自动更新配置；MyGO 更新签名密钥应通过环境变量提供，不要提交到仓库。

## 安全

- 订阅源内容不可信。正文必须在 Go 端经过 `internal/sanitize` 的 bluemonday 策略清洗后
  入库，并禁用 `script`、`style`、`iframe`、`object`、`embed`、`form`、`input`、
  `button` 以及 `style`、`srcset`、`onerror`、`onload`。它还会把链接改为绝对地址并加
  `target="_blank" rel="noopener noreferrer"`，图片改为懒加载。
- `post-view.tsx` 和快照卡片里的 `innerHTML` 只允许接收数据库中已清洗的正文；不要把抓取
  原文直接传给前端。嵌入 WebView 的快照必须通过 `Page.WebView().CapturePage()` 捕获。
- 打开外链前必须经过 `externalHref()`，只允许 `http:`/`https:`。`Settings.OpenLink`
  再按偏好调用系统浏览器或创建应用内窗口；不要把原始 href 直接交给 shell 或窗口。
- 订阅地址在 Go 侧由 `rss.ValidateURL` 再校验一次；不要只依赖前端校验。
- 数据是本地未加密数据库，包含订阅、完整正文、已读/星标和偏好；设置对话框可打开数据目录。
- 没有账号、同步、遥测或统计。WebView 不直接加载订阅源图标；图标由 Go 缓存后以本地 data
  URL 渲染。正文远端图片默认由 Go 在 `Posts.Get` 中改写为 `ressu-media://`，通过本地缓存协议
  提供；用户可在设置中选择 Webview 直接加载。
- webview 端禁止直接网络请求，由 `.oxlintrc.json` 的 `no-restricted-globals` 和
  `webview/test/setup.ts` 的网络 stub 双重执行。需要远程数据时新增 Go 绑定。
- `mygo build` 从 `mygo://localhost` 提供内嵌页面；页面没有 CSP meta 标签，因此 XSS 防护
  依赖内容清洗和 URL 校验，而不是 CSP。
- 快照保存路径来自原生保存对话框；文件名由文章标题清理生成，只写入用户选中的路径。
  不要把快照上传到网络或写入订阅源可控的路径。
- 不要提交凭据、用户数据或生成物。签名/更新密钥放在环境变量中。
