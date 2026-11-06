# Ressu

一个免费的现代 **RSS 阅读器**，桌面端用 [MyGO](https://github.com/egoist/mygo)（纯 Go、无 cgo、WebView），前端用 SolidJS。

[![Listed on Awesome MyGo](https://awesomemygo.dev/badge.svg)](https://awesomemygo.dev/apps/ressu)

> 主要面向 Windows，其次是 Linux 与 macOS。

<img width="2009" height="1370" alt="图片" src="https://github.com/user-attachments/assets/b3f799c9-6a9b-425b-a331-37f197d797ff" />

## 技术栈

| 层       | 选择                                                         |
| -------- | ------------------------------------------------------------ |
| 桌面框架 | MyGO 0.7（WebView2 / WebKitGTK / WKWebView）                 |
| 后端     | Go 1.27，`gofeed` 解析 RSS 0.9x / 1.0 / 2.0、Atom、JSON Feed |
| 前端     | SolidJS 1.9 + solid-ui（Kobalte）+ Tailwind CSS 4 + Vite 8   |
| 状态     | Solid store（不再需要 Redux）                                |
| 工具链   | Bun、oxlint、oxfmt、Vitest、TypeScript                       |
| 存储     | 用户数据目录下的 bbolt 数据库（`ressu.db`）                  |

抓取、解析、去重、存储全部在 Go 侧完成，页面通过 `mygo generate` 生成的类型化客户端调用，刷新进度用类型化事件回推。

## 开发

需要：Go 1.27+、Bun、以及系统 WebView（Windows 10 需 WebView2 Runtime，Windows 11 自带）。

```sh
bun install
bun run dev        # mygo dev：编译并启动应用，同时启动 Vite 热更新
```

前端源码与构建配置都在 `webview/`，但所有脚本仍在仓库根目录执行（Vite 的 root 指回 `webview/`）。

常用脚本：

| 命令               | 作用                                                                 |
| ------------------ | -------------------------------------------------------------------- |
| `bun run dev`      | 启动应用（Go + Vite）                                                |
| `bun run dev:web`  | 只启动 Vite，在普通浏览器里调布局（此时读不到数据）                  |
| `bun run generate` | 重新生成 `webview/src/mygo.ts`（`mygo dev`/`build` 会自动生成）      |
| `bun run check`    | `tsc -p webview/tsconfig.json --noEmit` + `oxlint` + `oxfmt --check` |
| `bun run fmt`      | 用 oxfmt 格式化（含 import 与 Tailwind 类名排序）                    |
| `bun run test`     | Vitest（jsdom + @solidjs/testing-library）                           |
| `go test ./...`    | Go 单测与集成测试（本地 httptest 服务器跑完整的抓取流程）            |
| `bun run build`    | 打包应用，产物在 `build/<os>-<arch>/`                                |

## 目录结构

```
main.go                 应用入口：绑定服务、菜单（仅 macOS）、窗口、启动刷新
internal/
  rss/                  抓取（ETag / If-Modified-Since）与解析
  store/                bbolt 持久化：订阅源、文章、偏好都放在 ressu.db
  service/              暴露给前端的服务与事件（Feeds / Posts / Settings）
webview/                前端源码与构建配置（独立的源代码目录）
  index.html            Vite 入口
  vite.config.ts        Solid + Tailwind 插件，root 指向 webview/
  vitest.config.ts      测试配置
  src/
    mygo.ts             由 Go 生成的类型化客户端（git 忽略）
    app.tsx             三栏布局与启动流程
    stores/             应用状态、事件订阅
    components/         侧栏、文章列表、阅读视图、设置、对话框
    components/ui/      solid-ui 组件（shadcn 风格的源码内联）
    lib/                日期格式、HTML 清洗、主题
  test/setup.ts         Vitest 的 jsdom 垫片
  public/               Vite 公共资源
resources/icon.png      应用图标
```

## 数据放在哪

`Ressu` 的 Go 侧把数据写在系统的用户数据目录（Windows 上是
`%AppData%\Ressu`），设置 → 数据 里可以直接打开该目录：

- `ressu.db`：bbolt 数据库，含订阅源与未读数（`feeds` 桶）、文章（`posts` 桶，每个订阅源
  一个嵌套桶，默认每源保留 500 篇，可在设置里调整）、偏好（`settings` 桶）
- `window-state.json`：窗口位置与大小

数据库由 bbolt 负责事务与崩溃恢复，写入不会留下半个文件。同一时间只能有一个实例打开它。

## 功能

- 订阅源：添加（RSS / Atom / JSON Feed）、取消订阅、搜索、右键菜单（刷新、全部已读、取消订阅）、未读徽标
- 抓取：启动时刷新、手动刷新单个或全部（4 路并发），带 ETag 缓存与逐源错误提示
- 阅读：三栏可拖拽布局、未读/星标筛选、打开即已读、标记未读、星标、正文 HTML 清洗后渲染
- 设置：跟随系统/浅色/深色主题、启动刷新、打开链接方式（系统浏览器或应用内窗口）、每源保留条数、数据目录
- 界面：暗色模式、同步进度状态栏、Toast 提示、键盘可达的列表项

## 后续

- 「全部文章」汇总视图
- 订阅分组
- 单栏 / 双栏布局模式
- 全文搜索（FTS）
- 自动更新

## LICENSE

This is free and unencumbered software released into the public domain.
See [LICENSE](LICENSE).
