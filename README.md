# Ressu

一个随处可见，普通，没有什么特点的 **RSS 阅读器**，使用 [MyGO](https://github.com/egoist/mygo) + solid.js 构建，基于Webview的显示效果。

<img width="1229" height="823" alt="图片" src="https://github.com/user-attachments/assets/6df8c8af-92b2-4ee6-82ea-ddd085944aa4" />

## 功能

- 支持添加（RSS / Atom / JSON Feed）
- 从 OPML 批量导入订阅源，或在设置里导出
- 没有固定的抓取间隔设置，无通知，只有在你想用的时候，回到前台会自动刷新
- 三栏/二栏可拖拽布局
- 支持文章生成完整长图，方便和群友分享
- 跟随系统/浅色/深色主题

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
