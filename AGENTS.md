# AGENTS.md

Ressu 是一个免费的桌面 RSS 阅读器。所有数据由一个 Go 进程持有：它抓取并解析订阅源、按 GUID 去重、把文章存进用户数据目录下的 bbolt 数据库，再通过 MyGO 绑定把数据交给 SolidJS 页面。页面只负责渲染，自己从不为了订阅数据访问网络。

本文件说明仓库如何构建、组织和改动。面向用户的中文介绍在 `README.md`。

## 技术栈

| 层     | 选择                                                                    |
| ------ | ----------------------------------------------------------------------- |
| 外壳   | MyGO 0.7（`github.com/egoist/mygo`），Go，无 cgo，系统 WebView          |
| 后端   | Go 1.27.3，`mmcdole/gofeed` 解析 RSS 0.9x/1.0/2.0、Atom、JSON Feed      |
| 前端   | SolidJS 1.9，shadcn 风格的 Kobalte + corvu 原语，Tailwind CSS 4         |
| 打包   | Vite 8 + `vite-plugin-solid`，构建目标 `es2022`                         |
| 状态   | 单个 Solid store（`webview/src/stores/app.ts`），没有 Redux，也没有路由 |
| 工具链 | Bun（锁文件 `bun.lock`）、TypeScript 7、oxlint、oxfmt、Vitest 5         |
| 存储   | MyGO 用户数据目录下的 bbolt 数据库（`ressu.db`）                        |
| 平台   | 以 Windows 为主；Linux 与 macOS 也支持，并在 CI 里打包                  |
| 许可证 | Unlicense（公有领域）                                                   |

仓库里没有 `pyproject.toml`，也没有 `Cargo.toml`。真正相关的两个清单是 `package.json`（前端脚本与依赖）和 `go.mod`（`module github.com/sdttttt/ressu`）。

## 目录结构

```
main.go                     应用入口：打开存储、绑定服务、窗口、启动刷新
menu_darwin.go              macOS：系统菜单及其命令
menu_other.go               Windows/Linux：SetMenu(nil) 移除原生菜单栏
internal/                   Go internal/ 规则下的后端包（外部不可导入）
  rss/                      HTTP 抓取（ETag/If-Modified-Since）与订阅源解析
    fetch.go, parse.go      Fetch/ValidateURL、Parse/Meta/Item、plainText/truncate
    testdata/               rss2.xml、atom.xml、jsonfeed.json、broken.xml 测试夹具
  store/                    bbolt 持久化，订阅源与偏好缓存在内存
    store.go, model.go
  service/                  暴露给页面的 API 与类型化事件
    feeds.go, posts.go, settings.go, events.go
webview/                    前端源码与构建配置（独立的源代码目录）
  index.html                Vite 入口（lang="zh-CN"，挂载 /src/main.tsx）
  vite.config.ts            Solid + Tailwind 插件、`~` 别名、root 指回 webview/、5173 端口
  vitest.config.ts          jsdom、setup 文件、`src/**/*.test.{ts,tsx}`
  tsconfig.json             strict、noEmit、jsx preserve（solid-js）、路径 `~/*`
  src/                      SolidJS 前端（别名 `~` 指向这里）
    main.tsx                把 <App/> 渲染进 #root，先按系统主题上色
    app.tsx                 三栏可拖拽布局、状态栏、对话框、toast
    mygo.ts                 生成的类型化客户端（已被 git 忽略；不要编辑）
    index.css               Tailwind 4 入口：主题变量、dark 变体、.post-body
    components/             feed-sidebar、post-list、post-view、*-dialog、status-bar、empty-state
    components/icon-button.tsx  solid-ui Button + Tooltip 的图标按钮（aria-label 即可访问名）
    components/ui/          源码内联的 shadcn 风格原语（Kobalte/corvu + cva）
    lib/                    format.ts（zh-CN Intl）、sanitize.ts（DOMPurify）、theme.ts、utils.ts
    stores/                 app.ts（状态 + 动作）、events.ts（事件订阅）
  test/setup.ts             Vitest 的 jsdom 垫片（matchMedia、ResizeObserver、scrollIntoView）
  public/favicon.svg        Vite 公共资源
  dist/                     前端构建输出（已被 git 忽略）
resources/icon.png          MyGO 打包使用的应用图标
mygo.config.ts              MyGO CLI 配置（name、slug、identifier、dev/build 命令、out）
.oxlintrc.json              oxlint 插件/规则与忽略项
.oxfmtrc.json               oxfmt：printWidth 100、import 排序、Tailwind 类名排序
.github/workflows/ci.yml    ubuntu/windows 测试矩阵，v* tag 时打包
build/                      构建产物（已被忽略）：Ressu.exe、安装包、图标
```

`webview/dist/`、`build/`、`.mygo/`、`webview/src/mygo.ts` 都是生成物，已被 git 忽略。

## 构建与测试命令

首次安装用 `bun install`（包管理器是 Bun；不要引入 npm 或 pnpm 的锁文件）。

| 命令                 | 作用                                                                                                                          |
| -------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `bun run dev`        | `mygo dev`：通过 `devCommand` 拉起 Vite，构建并启动 `.mygo/dev` 里的开发版应用                                                |
| `bun run dev:web`    | 只跑 `http://localhost:5173` 的 Vite；可调布局，但调 Go 会失败（没有数据）                                                    |
| `bun run generate`   | 依据 `mygo.Bind` 服务与 `mygo.NewEvent` 事件重新生成 `webview/src/mygo.ts`                                                    |
| `bun run typecheck`  | `tsc -p webview/tsconfig.json --noEmit`                                                                                       |
| `bun run lint`       | `oxlint`                                                                                                                      |
| `bun run fmt`        | `oxfmt`（会写文件；排序 import 与 Tailwind 类名）                                                                             |
| `bun run fmt:check`  | `oxfmt --check`                                                                                                               |
| `bun run check`      | typecheck + lint + fmt:check；CI 的关卡，**不**跑测试                                                                         |
| `bun run test`       | `vitest run`（jsdom + @solidjs/testing-library）                                                                              |
| `bun run test:watch` | Vitest 监听模式                                                                                                               |
| `go test ./...`      | Go 单测与集成测试（httptest 提供真实订阅源）                                                                                  |
| `bun run build`      | `mygo build`：先 `tsc -p webview/tsconfig.json --noEmit && vite build --config webview/vite.config.ts`，再编译打包到 `build/` |
| `bun run build:web`  | 只构建前端：类型检查后 `vite build` 输出到 `webview/dist/`                                                                    |
| `bunx mygo vet`      | MyGO 自带的 Go vet + 原生 UI 生命周期检查；未接入脚本或 CI                                                                    |
| `bunx mygo doctor`   | 检查开发环境是否就绪                                                                                                          |

注意：

- `webview/src/mygo.ts` 是生成物且被 git 忽略。全新克隆后要先 `bun run generate` 再
  `bun run check`，否则类型检查会因模块缺失而失败。`mygo dev` 与 `mygo build`
  会自己重新生成它。
- `bun run build` 默认 `-platform windows/amd64`；CI 会显式传
  `-platform windows/amd64`、`linux/amd64`、`darwin/universal`。产物落在
  `build/<os>-<arch>/`（Windows 下是 `Ressu.exe`、
  `ressu-setup-<version>-amd64.exe`、`icon.png`）。
- `mygo dev` 监听 Go 代码变化并重建/重启开发版应用；前端改动交给 Vite。该终端里
  `r` 重建、`c` 清屏、`q` 退出。
- 环境要求：Go 1.27.3+（来自 `go.mod`）、Bun，以及系统 WebView（Windows 10
  需要 WebView2 Runtime，Windows 11 自带）。Linux 需要 WebKitGTK 4.1；CI 会安装
  `libwebkit2gtk-4.1-0`、`libayatana-appindicator3-1`（确切包名见
  `.github/workflows/ci.yml`）。
- 开发服务器端口固定为 5173（`strictPort`），必须与 `mygo.config.ts` 里的
  `devUrl` 保持一致。

在本次检出上验证过（Windows、Go 1.27.3、Bun 1.4.4）：`go test ./...`、
`bun run check`、`bun run test`（4 个文件 34 个用例）、`bun run build:web`、
`gofmt -l .`（干净）、`go vet ./...` 与 `bunx mygo vet` 全部通过。

## 运行时架构

启动流程（`main.go`）：

1. `mygo.App.SetName("Ressu")`，解析 `mygo.PathUserData`，`store.Open(dir)`。
2. `mygo.Bind(service.NewFeeds(st), service.NewPosts(st), service.NewSettings(st))`。
3. `WhenReady` 里：用 `mygo.Theme.SetSource` 应用保存的主题、在 macOS 上安装系统菜单
   （Windows/Linux 调用 `SetMenu(nil)` 移除原生菜单栏）、打开主窗口
   （1200x800，最小 900x600，`StateKey: "main"`，因此 MyGO 会写
   `window-state.json`）。窗口以 `Hidden` 创建，在 `OnReadyToShow` 里才 `Show`，所以不会
   出现空白窗口；5 秒内页面仍未就绪也会兜底显示。`refreshOnStart` 的刷新在这个回调里
   启动，避免和首帧抢资源。
4. `app.Run()` 阻塞；启动阶段的错误走 `log.Fatal`。

绑定与事件：

- 绑定服务上任何导出方法都能被页面调用。首参 `context.Context` 由 MyGO 注入，并在
  TypeScript 中省略（`Feeds.Add(ctx, url)` 变成 `Feeds.add(url)`）。参数与返回值必须是
  可 JSON 序列化的 struct/slice/string/number/bool/`time.Time`；可空字段用指针
  （`*time.Time` -> `string | null`）。
- `mygo generate` 写出 `webview/src/mygo.ts`：从 Go struct 复制的接口（JSON tag 转 camelCase、
  保留文档注释）、调用 `call("Service.Method", ...)` 的 `Feeds`/`Posts`/`Settings`
  对象，以及由 `mygo.NewEvent[T](name)` 生成的 `events.*`。调用失败时以 Go 错误文本
  reject。
- 事件名用冒号分区：`sync:state`、`feeds:changed`、`posts:changed`、`ui:command`
  （声明在 `internal/service/events.go`）。
- 原生菜单只存在于 macOS，它不直接拥有对话框：它广播 `ui:command` 字符串（`add-feed`、
  `refresh`、`settings`，常量见 `service.Command*`），由 `webview/src/stores/events.ts`
  打开对话框或执行动作。Windows/Linux 没有菜单栏，页面在 `app.tsx` 里注册同样的
  `Ctrl/Cmd+N`、`Ctrl/Cmd+R`、`Ctrl/Cmd+,` 快捷键。

数据流：

- 页面启动（`webview/src/app.tsx` 的 onMount -> `start()`）：订阅事件、监听系统主题，然后并行
  加载 `Feeds.list()`、`Settings.get()`、`Settings.dataDir()`，应用主题，选中第一个
  订阅源。选中订阅源会加载它第 0 页的文章。
- 刷新：`refreshAll()` -> Go 的 `Feeds.RefreshAll`，单飞（`ErrRefreshing`），广播带
  `active/total` 的 `sync:state`，并以信号量每次抓取 4 个订阅源
  （`maxConcurrent = 4`）。成功的 `refreshOne` 会保存新的校验器和合并后的文章，并广播
  `feeds:changed` 与 `posts:changed`；抓取失败只广播 `feeds:changed`。`RefreshAll`
  在每个订阅源完成后广播一次 `sync:state`，最后再广播一次 `active: false`。
- 304（`rss.ErrNotModified`）保留已存文章，只清除错误并更新 `lastFetchedAt`。失败则把
  `lastError` 记在订阅源上，保留订阅源与文章，并在状态栏和 toast 里展示错误文本。
- 任何写操作都会广播 `feeds:changed`（见 `broadcastFeeds`）；新增会改状态的服务方法也应
  照做，这样多窗口才能保持一致。
- 筛选在服务端完成：`Posts.List` 接收
  `{feedId, unreadOnly, starredOnly, search, offset, limit}`，返回摘要加 `total` 和
  `unread`。界面每页 100 条（`PAGE_SIZE`），服务默认 100、上限 500。`search` 在 Go 里
  已实现，但当前界面始终传 `""`。
- 未读数由文章推导，存在 `Feed.Unread` 上：合并文章时重算，单条已读/取消已读时增减，
  全部已读时清零。

持久化（`internal/store`）：

数据库是用户数据目录下的 `ressu.db`（bbolt，单文件）：

| 桶         | 键                  | 值                                                          |
| ---------- | ------------------- | ----------------------------------------------------------- |
| `feeds`    | 8 字节大端订阅源 id | 整个 `Feed` 的 JSON：频道元数据、ETag/Last-Modified、未读数 |
| `posts`    | 订阅源 id（嵌套桶） | 每个 GUID 一条 `Post` 的 JSON                               |
| `settings` | `prefs`             | 整个 `Prefs` 的 JSON                                        |

`window-state.json` 仍由 MyGO 依据 `StateKey` 写在同一个目录，不是本仓库代码。

- `Store` 把订阅源和偏好缓存在内存（数量小、读频繁），文章每次从数据库读；一把互斥锁串行化
  写操作。出参都做克隆（`slices.Clone`），内部值不会被调用方改到。
- 只有一个进程能用数据库：bbolt 的文件锁会让第二个实例在打开时超时报错。
- 没有版本键，也没有迁移代码。改结构时要自己兼容旧记录，或加一个版本键并写迁移。
- `MergePosts` 按 GUID 刷新已存在的条目但保留 `read`、`starred`、`fetchedAt`；按
  `publishedAt` 倒序；裁剪到 `prefs.PostsPerFeed`（默认 500，允许 50–2000）并重算未读。
  它在一个事务里重写该订阅源的整个文章桶，单条已读/星标只写那一条。
- `AddFeed` 用「最大 id + 1」，删除最高 id 的订阅源后该 id 会被复用。
- `Prefs.Normalize` 替换越界值；未知主题回落到 `system`，`openLinks` 不是 `window` 就
  变成 `browser`。前端在 `webview/src/stores/app.ts` 的 `defaultPrefs` 里复制了一份默认值——
  新增偏好时两处都要改。
- 文章的 `content` 是订阅源原始 HTML；`summary` 是截到 280 个字符加省略号的纯文本。
  HTML 只在渲染时清洗。

订阅源抓取（`internal/rss`）：

- 只有绝对的 `http`/`https` 地址能通过 `ValidateURL`，其它一律返回 `ErrInvalidURL`。
- `Fetch` 发送 `User-Agent: Ressu/0.1 (+https://github.com/sdttttt/ressu)`、偏向订阅源的
  `Accept`，以及 `If-None-Match`/`If-Modified-Since`。请求超时 30s（`Timeout`），响应体
  上限 16 MiB（`MaxBody`），304 映射为 `ErrNotModified`。
- `Parse` 用链接补齐缺失的 GUID，再退回标题+日期；拒绝不是订阅源的页面；把相对链接按
  订阅源地址解析为绝对地址；摘要统一成单行纯文本。

前端：

- `webview/src/stores/app.ts` 是一个 Solid store；读取时导入 `app`，修改时用导出的函数
  （`start`、`selectFeed`、`loadPosts`、`selectPost`、`refreshAll`、`addFeed`、
  `removeFeed`、`updatePrefs` 等）。组件从不直接调用 `~/mygo`（测试里 mock 它除外）。
- 动作会捕获错误并用 `messageOf(error)` 弹中文 toast；只有调用方需要按成功与否分支时才
  返回布尔值（`addFeed`）。事件处理里的“发后不管”用 `void action()`。
- 派生值是普通函数（`currentFeed()`、`visibleFeeds()`、`totalUnread()`、
  `isSyncingFeed(id)`），不是 memo。
- 事件在 `App` 里通过 `webview/src/stores/events.ts` 的 `subscribe()` 只订阅一次；它返回的退订
  函数用在 `onCleanup` 里。
- 主题应用两次：Go 设置原生窗口外观，`webview/src/lib/theme.ts` 切换 `<html>` 上的 `dark` 类
  （Tailwind 的 dark 变体是 `&:where(.dark, .dark *)`）。CSS 变量是 `webview/src/index.css` 里的
  OKLCH 值，并使用 MyGO 提供的 `--mygo-accent` 与 `--mygo-text-scale`。
- 没有路由：只有一页，对话框由 `app.dialog` 驱动。
- 文章 HTML 用 `innerHTML={sanitizeHTML(...)}` 渲染；见“安全”一节。
- `Intl.DateTimeFormat("zh-CN")` 放在 `webview/src/lib/format.ts`；日期函数接受可选的 `now`
  参数，让测试保持确定性。

## 代码约定

语言：代码注释、文档注释、标识符与提交信息用英文；面向用户的文案（界面字符串、toast、
README、`lang="zh-CN"`）用简体中文。本文件（AGENTS.md）是例外，用中文书写。新界面文案
保持中文，除非任务明确要求，不要引入 i18n 层。

Go：

- 包的注释写在主文件上；每个导出标识符都要有文档注释，写成以该标识符开头的完整句子。
- 哨兵错误是包级 `errors.New` 值（`store.ErrNotFound`、`store.ErrDuplicate`、
  `rss.ErrNotModified`、`service.ErrRefreshing`）；用 `%w` 包装并带上文件上下文。错误
  信息要短且用户可读，因为页面会把它们原样显示在中文 toast 里。
- 每个持久化/序列化字段都要有 camelCase 的 `json` tag。
- 跟随标准库风格：`slices`、`min`/`max`、合适时用泛型、`for i := range n`。`gofmt`
  必须干净；改过的文件跑 `gofmt -w`。
- 平台差异用 `//go:build` 构建标签拆到不同文件（如 `menu_darwin.go` / `menu_other.go`），
  不要在代码里判断 `runtime.GOOS`。
- `Store` 上的方法保持加锁，出参克隆切片。
- 测试与被测代码同目录，命名为 `*_test.go`；不使用 mock 库——用 `t.TempDir()` 与
  `net/http/httptest`。

TypeScript / Solid：

- 文件与目录用 kebab-case；组件用 PascalCase，放在 `webview/src/components/<name>.tsx` 并默认
  导出。小的内部片段（`PostRow`、`FeedIcon`、`ChoiceGroup`）留在同一文件里，不导出。
- 从 `~/*` 导入；oxfmt 会排序 import 和 Tailwind 类名，所以直接跑 `bun run fmt`，不要
  手工对齐。
- Props 用内联对象类型；类名覆盖通过 `class` 加 `cn(...)`。变体用 `cva`（参考
  `components/ui/button.tsx`）。
- 只用语义化 token：`bg-background`、`text-muted-foreground`、`bg-accent`、
  `border-border`、`text-warning`、`bg-error`、`rounded-md` 等。不要写死颜色。Tailwind 4
  是 CSS-first 的：主题值在 `webview/src/index.css`（`@theme`）里，没有 `tailwind.config.js`。
- 图标来自 `lucide-solid`；toast 来自 `solid-sonner`（`toast`）。
- 界面控件一律用 `webview/src/components/ui/**` 里的 solid-ui 组件，不要再手写原生
  `input`/`select`/复选框：输入用 `text-field`、数字用 `number-field`、单选用
  `radio-group`、多选切换用 `toggle-group`、提示用 `tooltip`、计数用 `badge`、分隔用
  `separator`。图标按钮统一用 `~/components/icon-button`（solid-ui `Button` + `Tooltip`，
  `aria-label` 同时是提示文案与可访问名）。
- `webview/src/components/ui/**` 是源码内联的 shadcn 风格组件。保持它的 API 稳定；
  `.oxlintrc.json` 已经在这里放宽了部分 a11y/import 规则。
- 不要手改 `webview/src/mygo.ts`、`webview/dist/`、`build/`；重新生成或重新构建。
- 不要新增 ESLint/Prettier 配置——lint 是 oxlint，格式化是 oxfmt。

提交遵循 Conventional Commits（`feat:`、`fix:`、`chore:`，可带作用域如 `feat(ui):`，
破坏性改动加 `!`）。LF 行尾由 `.gitattributes` 强制。

### 改动清单

- 新增后端调用：在 `Feeds`、`Posts` 或 `Settings` 上加导出方法，广播它改动的状态，跑
  `bun run generate`，然后在 `webview/src/stores/app.ts` 的 store 动作里调用它（错误弹 toast、
  更新本地状态）。
- 新增偏好：在 `store.Prefs` 加带 JSON tag 的字段、在 `DefaultPrefs` 加默认值、在
  `Prefs.Normalize` 加范围限制、同步 `webview/src/stores/app.ts` 里的 `defaultPrefs`，并在
  `settings-dialog.tsx` 加控件。
- 新增事件：用 `mygo.NewEvent[T]` 声明，在状态变化处广播，重新生成，并在
  `webview/src/stores/events.ts` 订阅。
- 新增界面文案：中文、语义化颜色 token、`~` 导入、默认导出。

## 测试

- Go：`go test ./...`。三个包都有测试。`rss` 是表驱动的，夹具在
  `internal/rss/testdata/`（RSS 2.0、Atom、JSON Feed、坏 XML），覆盖
  `plainText`/`ValidateURL`；`store` 覆盖数据库读写、默认值、重复拒绝、合并语义、裁剪、
  全部已读、删除与重开；`service` 用 `httptest` 服务器跑完整的
  添加 -> 刷新 -> 304 -> 新增条目 -> 标记已读 -> 筛选流程，另有失败、分页与搜索的用例。
  测试夹具中英混排是有意为之。
- 前端：`bun run test`。Vitest + jsdom，垫片在 `webview/test/setup.ts`。测试用 `vi.mock` 和固定
  夹具 mock `~/mygo` 与/或 `~/stores/app`（mock 里用到的值用 `vi.hoisted` 提升）；
  `restoreMocks: true` 已开启。新测试放在 `webview/src/**/*.test.{ts,tsx}` 旁边。组件测试使用
  `@solidjs/testing-library` 的 `render`/`fireEvent`。
- 没有 E2E 套件、没有覆盖率阈值、没有浏览器自动化。测试不得访问网络，也不应依赖打包好
  的应用。
- 即使全部通过，store/组件测试仍会向 stderr 打印
  `You appear to have multiple instances of Solid`；它不会让测试失败。

## CI 与发布

`.github/workflows/ci.yml`：

- `test` 任务在推送到 `master`/`develop` 与 `v*` tag、以及向 `master`/`develop` 提 PR
  时运行。矩阵为 `ubuntu-latest` + `windows-latest`。步骤：
  `bun install --frozen-lockfile`、`go test ./...`、`bun run generate`、
  `bun run check`、`bun run test`、`bun run build:web`。
- `package` 任务只在 `v*` tag 上运行，`needs: test`，矩阵为 `windows/amd64`、
  `linux/amd64`、`darwin/universal`：`bun run build -- -platform <target>`，然后把
  `build/` 作为 artifact 上传。Linux 会先安装 WebKitGTK/AppIndicator 相关包。macOS 的
  bundle 与 DMG 需要 macOS；Windows 构建还会产出 NSIS 安装包。
- 发版时同时改 `package.json` 与 `mygo.config.ts` 里的版本号；tag 驱动打包。本仓库目前
  还没有自动更新配置。

## 安全注意事项

- 订阅源内容不可信。任何以 HTML 渲染的内容都必须经过 `sanitizeHTML`
  （`webview/src/lib/sanitize.ts`）：DOMPurify 的 HTML profile，并移除
  `script`、`style`、`iframe`、`object`、`embed`、`form`、`input`、`button` 以及
  `style`/`srcset`/`onerror`/`onload`。它还会把链接重写为绝对地址并加
  `target="_blank" rel="noopener noreferrer"`，把图片改成无固定尺寸的懒加载。`post-view.tsx`
  里的 `innerHTML` 调用必须始终走这个函数。
- 只有在 `externalHref()` 确认是 `http`/`https` 后才会打开链接。`Settings.OpenLink`
  再按偏好调用系统浏览器（`mygo.Shell.OpenExternal`）或打开应用内窗口。永远不要把原始
  href 直接交给 shell 或窗口。
- 订阅源地址在服务端校验（`rss.ValidateURL`）：只接受绝对的 `http`/`https`，因此
  `file://`、`javascript:` 和相对地址都会被拒。抓取有 30s 超时与 16 MiB 体积上限。
- 数据是本地未加密的数据库（`ressu.db`）：订阅、完整正文、已读与星标状态、偏好，外加
  MyGO 的 `window-state.json`。设置对话框会展示数据目录并可在文件管理器中打开。删除订阅源
  会删掉它在 `posts` 桶里的嵌套桶。
- 没有账号、同步、遥测或统计。文章引用的远程资源与订阅源图标由 WebView 直接从对方服务器
  加载，因此那些服务器能看到请求；这是除抓取订阅源之外唯一的出站流量。
- `mygo build` 会把前端内嵌并从 `mygo://localhost` 提供；页面没有
  Content-Security-Policy meta 标签，所以防 XSS 依赖内容清洗与 URL 校验，而不是 CSP。
- 不要提交生成物或用户数据（`build/`、`webview/dist/`、`.mygo/`、`webview/src/mygo.ts` 已被忽略）。
  不要把凭据写进 `mygo.config.ts` 或仓库；签名/更新密钥应放在环境变量里（MyGO 自动更新
  用 `MYGO_UPDATER_PRIVATE_KEY`，本仓库尚未使用）。
