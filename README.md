# Ressu

一个现代的 **RSS 阅读器**，桌面端用 [MyGO](https://github.com/egoist/mygo)（纯 Go、无 cgo、WebView），前端用 solid.js + solid-ui 构建。

主要面向 Windows，其次是 macOS 与 Linux。

[![Listed on Awesome MyGo](https://awesomemygo.dev/badge.svg)](https://awesomemygo.dev/apps/ressu)

<img width="2009" height="1370" alt="图片" src="https://github.com/user-attachments/assets/b3f799c9-6a9b-425b-a331-37f197d797ff" />

## 功能

- 订阅源：添加（RSS / Atom / JSON Feed）、取消订阅、搜索、右键菜单（刷新、全部已读、取消订阅）、未读徽标
- 抓取：启动时刷新、手动刷新单个或全部（4 路并发），带 ETag 缓存与逐源错误提示
- 阅读：三栏可拖拽布局、未读/星标筛选、打开即已读、标记未读、星标、正文 HTML 清洗后渲染
- 设置：跟随系统/浅色/深色主题、启动刷新、打开链接方式（系统浏览器或应用内窗口）、每源保留条数、数据目录
- 界面：暗色模式、同步进度状态栏、Toast 提示、键盘可达的列表项

## 数据放在哪

`Ressu` 的 Go 侧把数据写在系统的用户数据目录（Windows 上是
`%AppData%\Ressu`），设置 → 数据 里可以直接打开该目录：

- `ressu.db`：bbolt 数据库，含订阅源与未读数（`feeds` 桶）、文章（`posts` 桶，每个订阅源
  一个嵌套桶，默认每源保留 500 篇，可在设置里调整）、偏好（`settings` 桶）
- `window-state.json`：窗口位置与大小

## 开发

需要：Go 1.27+、Bun、以及系统 WebView（Windows 10 需 WebView2 Runtime，Windows 11 自带）。

```sh
bun install
bun run dev        # mygo dev：编译并启动应用，同时启动 Vite 热更新
```

前端源码与构建配置都在 `webview/`，但所有脚本仍在仓库根目录执行（Vite 的 root 指回 `webview/`）。

## LICENSE

See [LICENSE](LICENSE) (0BSD).
