# Ressu 阅读页前端（webview）

Vite + SolidJS + TypeScript（最小依赖），Bun 管理依赖，产物由 MyGO 内嵌进系统
WebView。页面**没有网络权限**：所有数据走生成的类型化绑定（`src/mygo.ts`），事件来自
唯一的 `app:event` 通道。开发期 vite dev server 的 HMR 是唯一例外，产物代码不得包含
任何网络 API 调用。

## 目录结构

| 路径                 | 职责                                                       |
| -------------------- | ---------------------------------------------------------- |
| `index.html`         | 阅读页入口（仅挂载点），`lang="zh-CN"`                      |
| `snapshot.html`      | 快照采集页入口，独立子窗口加载                              |
| `vite.config.ts`     | 多页构建（两个入口）、dev server 端口 1886                  |
| `vitest.config.ts`   | vitest 配置（node 环境，跑 `src/**/*.test.ts`）             |
| `tsconfig.json`      | 严格模式 TS 配置（`tsc --noEmit` 即 `bun run typecheck`）   |
| `src/mygo.ts`        | `bun run generate` 生成的类型化绑定，**不要手改**           |
| `src/reader/`        | 阅读页（Solid 组件），见下                                  |
| `src/snapshot/`      | 快照采集页（Solid 组件），见下                              |
| `dist/`              | vite 产物（生成物，git 忽略，`mygo.config.ts` 的 `frontendDist`）|

## 构建与开发

- `bun run dev` 按 `mygo.config.ts` 的 `devCommand` 自动启动 vite dev server（HMR），
  窗口加载 `devUrl`（http://localhost:1886）。
- `bun run build` 按 `buildCommand` 先构建前端再内嵌打包。
- `bun run typecheck` 类型检查；`bun run test` 跑 vitest；`bun run check` 两者一起；
  `bun run build:web` 只构建前端（调试产物用）。

## reader 页（src/reader/）

| 文件                 | 职责                                                       |
| -------------------- | ---------------------------------------------------------- |
| `main.tsx`           | 入口：mygo 守卫、桌面外观变量、挂载 `<ReaderApp/>`          |
| `store.ts`           | 大脑：prefs/post/层栈 signal、事件 dispatcher、层动画引擎、进度条 |
| `app.tsx`            | 布局（空态/层栈/进度条）+ 生命周期接线（事件、快捷键、点击委托、resize） |
| `article-layer.tsx`  | 单个文章层：骨架、正文 innerHTML、kind effect 驱动滑动      |
| `attachments.tsx`    | 附件条组件                                                  |
| `links.ts` / `enclosures.ts` | 纯函数（外链过滤、附件名/大小），配 vitest 用例     |
| `styles.css`         | 排版、主题 token、层与进度条样式                            |

- 状态集中在 `store.ts`：`prefs`/`post`/`layers` 三个 signal；事件 revision 与请求
  requestId 是非响应式计数器。dispatcher 按 `eid` 路由，丢弃 `revision` 不递增的事件。
  消费分支：`state.snapshot`/`appearance.changed`/`prefs.changed`、`selection.changed`
  （清空文章栈）、`post.selected`（打开正文）、`post.fulltext`（匹配当前文章时替换正文）。
- **文章层栈**：`Layer` 类持有 `kind`/`post` 两个 signal 与非响应式字段（element、
  animation、swap）；`<For>` 按对象身份复用 DOM，被盖住的层保留滚动位置。换篇动画由
  原生列表选择方向决定（`direction`：1 新文章从下方淡入、旧层上滑退出；-1 新层从上方
  盖下）。动画只在元素挂载后的下一帧启动（先内联钉住首帧），用 Web Animations API，
  被替换的层有 `finished` 回调 + 超时兜底回收；减少动效偏好或系统
  `prefers-reduced-motion` 下跳过滑动直接进终态，正文翻页的平滑滚动同步降级。
- 正文经 `.post-content` 包装进 `innerHTML`（已清洗），使附件条保持为 `.post-body` 的
  直接子元素；CSS 的间距规则同时覆盖 `.post-body > * + *` 与 `.post-content > * + *`。
- **阅读进度条**：`#progress` 吸顶细条（`muted-foreground` 混色灰，随三套主题自适应）
  跟踪顶层文章层的滚动进度，可见更新经 CSS transition 缓动逼近，换篇重现时瞬时到位，
  减少动效时全程无动画；scroll 事件不冒泡，用 `#article` 上一处 capture 监听覆盖所有层；
  一屏读完的短文章隐藏。进度是度量驱动的，经 `registerProgress` 拿到的元素直接命令式
  更新，不走 signal。
- 打开正文由 `Posts.get` 完成；已读标记在 Go 侧按偏好自动处理（见 core/README），`m`
  键调 `Posts.toggleRead` 服务端翻转。
- 主题两层：`action.UpdatePreferences` 设置原生外观；页面在 `<html>` 上切换 `dark` 和
  `theme-sepia` 类，字号经 `--app-text-scale` 缩放。`system` 主题监听
  `prefers-color-scheme`。`sepia` 的运行时类名必须是 `theme-sepia`。桌面强调色与文字
  缩放经 `mygo-runtime` 的 `preferences()` 写入 `--mygo-accent`/`--mygo-text-scale`。

## snapshot 页（src/snapshot/）

- `app.tsx`：`<SnapshotApp/>` 拉取选中文章、渲染固定浅色卡片、按视口逐片滚动采集
  （HUD 间歇显示报进度），交给 `Posts.stitchSnapshot` 拼接；Esc/关闭按钮走
  `Snapshot.cancel`。`offsets.ts` 是纯函数 `captureOffsets`（1px 重叠、32 片上限），
  配 vitest 用例。
- 卡片标题、作者、链接等一律 Solid 插值（自动转义）；正文与 reader 一样经
  `.post-content` 进 `innerHTML`。

## 安全约定

- `Post.Content` 是 Go 端已清洗的 HTML，可直接进入正文容器；其他不可信文本用
  Solid `{}` 插值（自动转义）或 `textContent`。禁止
  `fetch`/`XMLHttpRequest`/`WebSocket`/`EventSource`。
- 正文里的链接点击先经 `externalHref()` 过滤只留 `http:`/`https:`，再交
  `Settings.openLink`；图片点击走 `Posts.openImage` 原生预览。
- 附件条地址再过一次 `externalHref()`。音频用原生 `<audio controls>` 播放，WebView
  直连附件地址，与正文内嵌媒体同语义（不经 Go 代理）；「下载附件」走
  `Posts.downloadEnclosure` Go 绑定，经安全下载管线后由原生保存对话框选路径，页面不
  发起任何下载请求。
- 样式使用 `styles.css` 的语义变量和系统主题类，不写死主题颜色；页面字符串用简体中文，
  TS 标识符和注释用英文。
