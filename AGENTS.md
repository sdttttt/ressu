# AGENTS.md

Ressu 是一个免费的桌面 RSS 阅读器：Go 进程抓取、解析、持久化订阅数据，经 MyGO 绑定交给
内嵌静态阅读页；订阅源栏、文章列表、文章头部和状态栏由 MyGO 原生 UI 渲染，页面只负责
阅读区正文。面向用户的介绍在 `README.md`。

本文件是仓库级开发说明（精简版）：只保留跨模块约定，**各子系统的细节已拆到包级 README，
动手前先读对应文档**：

| 主题                                 | 文档                       |
| ------------------------------------ | -------------------------- |
| 后端核心：抓取、持久化、刷新、事件   | `core/README.md`           |
| 应用门面与绑定分工                 | `core/app/README.md` |
| 原生界面：布局、侧栏、列表、窗口、快照 | `mygoui/README.md`  |
| 阅读页前端结构与安全约定             | `webview/static/README.md` |

## 快速约束

- 工具链只使用 Go：MyGO CLI 由 `go.mod` 的 `tool` 指令固定。不要引入 Bun、npm 或 pnpm 配置。
- Go 代码、标识符、代码注释、文档注释和提交信息用英文；界面字符串、toast、README 和
  `AGENTS.md` 用简体中文。
- 提交遵循 Conventional Commits，例如 `feat: ...`、`fix(ui): ...`、`docs: ...`。破坏性
  改动加 `!`。
- WebView 前端不得直接访问网络（`fetch`、`XMLHttpRequest`、`WebSocket`、`EventSource`）。
  新数据入口必须走 Go 绑定。
- 不要手改生成物：`build/`、`.mygo/`。
- 只改与任务相关的代码。不要顺手修复无关失败，也不要在任务中新增 ESLint/Prettier 配置。
- `gofmt` 必须干净（仓库带 `.golangci.yml`，启用 standard lint 集与 gofmt/goimports
  formatter）。

## 技术栈与模块

| 层     | 选择                                                              |
| ------ | ----------------------------------------------------------------- |
| 外壳   | MyGO 0.9.0（`github.com/egoist/mygo`），Go，无 cgo，系统 WebView  |
| 后端   | Go 1.27.3，`mmcdole/gofeed` 1.5，bbolt 1.5                        |
| 前端   | 无框架静态 HTML/CSS/JS，内嵌系统 WebView                            |
| 存储   | 用户数据目录下的 bbolt 单文件数据库 `ressu.db`                     |
| 平台   | 以 Windows 为主，同时支持 Linux 与 macOS                          |
| 许可证 | 0BSD                                                              |

清单是三个 Go module：根 `go.mod`（桌面壳）、`core/go.mod`（阅读器后端）和
`mygoui/go.mod`（MyGO 原生界面），由根目录 `go.work` 串成 workspace；各 go.mod 里的
`replace` 仅作脱离 workspace 单独构建的兜底。`core` 不得依赖 `mygo` 或 `mygoui`，保证
可以单独被 Web 服务、CLI 或移动端前端复用；`core` 只通过 `core/app` 对外暴露能力，
其余包全部在 `core/internal/` 下，禁止外部导入。

## 目录结构

```text
main.go                   极简启动器，只调用 mygoui.Run()
go.work                  串联根、core、mygoui 三个 module 的 workspace
core/                    独立后端 module，细节见 core/README.md
  app/                   唯一对外暴露的应用门面（入站端口）+ UI 接口、DTO 与
                         内部包重导出，细节见 core/app/README.md
  internal/              后端实现（rss/icon/sanitize/store/media/service/event/
                         status/logging/buildinfo/textmetrics/limitreader），
                         禁止 core 树以外的代码导入
mygoui/                  原生界面 module（启动装配、组合根与各功能包），细节见 mygoui/README.md
  run.go / menu_*.go     启动装配与 macOS 菜单
  webendpoint/           MyGO 绑定适配（Feeds/Posts/Settings/Snapshot/Status）
tools/pgo-collect/       在 CPU profile 下跑核心热路径，产出 default.pgo
tools/build/             一条命令完成 PGO 采样 + mygo 打包（发版入口）
webview/static/          阅读页与快照页静态源码，细节见 webview/static/README.md
resources/               打包图标
mygo.json                MyGO CLI 配置（应用名、版本号、内嵌前端目录、输出目录）
.github/workflows/       CI、打包与发布
default.pgo              PGO profile，go build 自动启用（-pgo=auto）
```

`build/`、`.mygo/` 和 `coverage/` 都是生成物或本地依赖，已被 git 忽略。

## 构建与测试

首次准备：`go mod download` 和 `go tool mygo doctor`。

| 命令                 | 作用                                                     |
| -------------------- | -------------------------------------------------------- |
| `go tool mygo dev`   | 构建并运行开发版；终端里 `r` 重建、`c` 清屏、`q` 退出     |
| `go test ./...`      | 壳 module 单测（`core/`、`mygoui/` 内同样命令跑各自 module） |
| `go vet ./...`       | Go 静态检查（三个 module 分别运行）                      |
| `go tool mygo build` | 内嵌静态前端、编译并打包到 `build/`                       |
| `go run ./tools/build` | 发版打包入口：先采样 PGO 再按 `-platform` 打包（缺省本机平台，`-skip-pgo` 复用现有 profile） |
| `go tool mygo vet`   | MyGO 的 Go vet 与原生 UI 生命周期检查；未接入 CI          |

注意：

- 运行中的 `mygo dev` 可能与 `mygo generate` 冲突并导致生成超时；重新生成前先退出开发版。
- 发布必须显式传 `-platform`；CI 分别构建 `windows/amd64`、`linux/amd64`、
  `darwin/universal`。Windows 10 需要 WebView2 Runtime；Linux 需要 WebKitGTK 4.1 与
  AppIndicator 相关包。
- PGO：`default.pgo` 随构建自动生效。`go run ./tools/build` 会先用
  `go run ./tools/pgo-collect -out default.pgo` 按当前代码重新采样；日常开发也可用
  `RESSU_CPUPROFILE=/tmp/ressu.pprof go tool mygo dev` 正常使用后用产出的 profile 替换；
  发版时 CI 的 `package` job 用同样方式重新采样。
- Go 集成测试绑定本地 httptest 端口，受限沙箱需允许本地监听。

## 架构总纲

数据自上而下走 `webview -> mygoui/webendpoint -> core/app -> service`，副作用自下而上由
`core/app` 回调给原生层；`core/internal/service` 只有业务逻辑，`core/app` 的 `App` 是界面能力的
唯一入口，`mygoui/webendpoint` 只做绑定适配。三者的完整契约见
`core/app/README.md`，事件 EID 清单与语义见 `core/README.md`，原生侧消费见
`mygoui/README.md`，页面侧消费见 `webview/static/README.md`。

几条跨层铁律：

- Go 只声明一个广播通道 `app:event`（载荷 `app.Event`，Hub 串行发布：revision →
  更新 `status.Model` 镜像 → sink 广播）。不要绕过 Hub，不要新增第二个通道。
- 事件分两类：`sync.*`、`feeds.changed`、`posts.changed`、`selection.changed`、
  `post.selected`、`post.updated` 是领域事件，任何前端都要消费；`appearance.changed`、
  `prefs.changed` 是视图事件，由桌面壳发布，移植其他前端可忽略。
- macOS 菜单项（设置、添加订阅、全部刷新）直接调用 `core/app` 的 `App` 方法，不经页面
  转发；Windows/Linux 没有原生菜单栏。
- 主窗口 1200x800、最小 900x600、`StateKey: "main"`，先 `Hidden` 创建，页面就绪
  （`Status.PageReady`）后显示，5 秒未就绪兜底显示。启动仅在 `RefreshOnStart` 为真时
  自动刷新，回前台经 `ForegroundRefresher` 触发 `RefreshStale`（5 分钟节流）。
- 快照入口在原生文章头部，经 `core/app` 的 `RequestSnapshot(parent)` 直接打开采集窗口。

## 代码约定

### Go

- 包注释写在主文件；每个导出标识符都要有以自身名字开头的文档注释。
- 哨兵错误是包级 `errors.New` 值（`store.ErrNotFound`、`store.ErrDuplicate`、
  `rss.ErrNotModified`、`rss.ErrInvalidURL`、`service.ErrRefreshing`）；需要上下文用
  `%w` 包装，错误文本要短且可直接展示在中文通知里。
- 每个持久化或跨绑定字段都要有 camelCase `json` tag。
- 跟随标准库风格：`slices`、`min`/`max`、泛型、`for i := range n`。
- 平台差异用 `//go:build` 拆文件，不要在代码里判断 `runtime.GOOS`。
- `Store` 方法保持加锁；返回切片时克隆。
- 测试与被测代码同目录，不用 mock 库，优先 `t.TempDir()`、`httptest` 和真实数据库。

### 静态 WebView

- 文件和目录用 kebab-case；不引入框架、打包器、UI 库或 CSS 框架。
- 页面字符串用简体中文；JavaScript 标识符和注释用英文。
- `Post.Content` 已清洗，可直接进正文容器；其他不可信文本用 `textContent`。新增交互
  先确认原生 UI 是否已承载该能力，只有阅读区能力才加入静态页。

## 常见改动路径

- 新增后端调用：`core/internal/service` 写业务逻辑（不碰 MyGO）→ `core/app` 加 `App` 方
  法（必要时封装 `mygoui` 方法并实现 `app.UI`，内部类型经 `export.go` 重导出）→
  `mygoui/webendpoint` 对应绑定对象加只做参数/调用方窗口适配的导出方法 → 页面或原生
  组件调用。细节见 `core/app/README.md`。
- 新增事件：`core/internal/event/event.go` 加 EID 和字段，补 `Validate` 与 Hub 发布方
  法，在 `reader.js` dispatcher 加分支。
- 新增偏好：`store.Prefs`、`DefaultPrefs`、`Prefs.Normalize`（在
  `core/internal/store`）→ `mygoui/settings` 加原生控件 → 若影响阅读更新 `reader.js`。
- 新增快捷键：先确认属于原生列表、主窗口还是静态正文，分别接在
  `mygoui/postlist`、`mygoui/app_view.go`、`reader.js`。WebView 焦点下转交
  原生行为的按键必须调用 `Status` 绑定并 `preventDefault()`。
- 新增主题：同步 Go 常量/校验、原生主题、CSS 变量、`reader.js`、原生设置窗口和测试。

## 测试

- Go：`go test ./...`（三个 module 分别在根目录、`core/`、`mygoui/` 跑）。`rss` 表驱动、
  夹具在 `core/internal/rss/testdata/`；`store` 覆盖创建/合并/裁剪/迁移/图标缓存；
  `service` 用 httptest 跑添加、刷新、304、节流、筛选、分页、搜索、OPML、已读；
  `event` 覆盖校验、revision 顺序、并发与快照；`core/app` 覆盖门面方法与 UI 副作用；
  `mygoui` 子包用 `ui.NewTester` 渲染各原生组件并断言行为（侧栏含选中胶囊滑动测试）。
- 静态页无独立测试框架；语法用 `node --check webview/static/reader.js` 检查。
- 没有 E2E、浏览器自动化或覆盖率阈值。提交前至少运行与改动相关的测试；跨层改动跑
  三个 module 的 `go test`/`go vet` 和 `go tool mygo build`。

## CI 与发布

`.github/workflows/ci.yml`：`test`（Ubuntu + Windows，跑测试、`node --check`、验证构建）→
`release-version`（校验 tag 为 `vYY.M.N` 且与 `mygo.json` 一致）→ `package`（三平台各自跑
`go run ./tools/build -platform`：先重新采样 PGO profile 再打包）→ `release`（汇总产物、生成
SHA256SUMS、发 GitHub Release；tag 含 `-` 标记预发布）。发版时修改 `mygo.json` 版本号并
打对应 tag。没有自动更新；签名密钥走环境变量，不进仓库。

## 安全

- 订阅源内容不可信：正文必须经 `core/internal/sanitize` 清洗后入库（禁用 script/iframe/form 等，
  链接加 `rel="noopener noreferrer"`）；前端 `innerHTML` 只接收已清洗正文。
- 外链必须经页面 `externalHref()` 与 `action.OpenLink` 双重校验，只允许绝对
  `http`/`https`。
- 快照只写入用户通过原生对话框选中的路径；数据是本地未加密数据库（设置窗口可打开数据
  目录）；没有账号、同步、遥测。
- 页面无 CSP，XSS 防护依赖清洗与 URL 校验；评审必须确认前端没有网络 API 调用。
- 不要提交凭据、用户数据或生成物。
