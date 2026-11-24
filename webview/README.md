# Ressu 阅读页前端（webview）

Vite + SolidJS + TypeScript（最小依赖），Bun 管理依赖，产物由 MyGO 内嵌进系统
WebView。页面**没有网络权限**：所有数据走生成的类型化绑定（`src/mygo.ts`），事件来自
唯一的 `app:event` 通道。开发期 vite dev server 的 HMR 是唯一例外，产物代码不得包含
任何网络 API 调用。

## 目录结构

| 路径                | 职责                                                       |
| ------------------- | ---------------------------------------------------------- |
| `index.html`        | 阅读页入口，`lang="zh-CN"`                                  |
| `snapshot.html`     | 快照采集页入口，独立子窗口加载                              |
| `vite.config.ts`    | 多页构建（两个入口）、dev server 端口 5173                  |
| `tsconfig.json`     | 严格模式 TS 配置（`tsc --noEmit` 即 `bun run typecheck`）   |
| `src/mygo.ts`       | `bun run generate` 生成的类型化绑定，**不要手改**      |
| `src/reader/`       | 阅读页：`main.ts` + `styles.css`                            |
| `src/snapshot/`     | 快照采集页：`main.ts` + `styles.css`                        |
| `dist/`             | vite 产物（生成物，git 忽略，`mygo.config.ts` 的 `frontendDist`）|

## 构建与开发

- `bun run dev` 按 `mygo.config.ts` 的 `devCommand` 自动启动 vite dev server（HMR），
  窗口加载 `devUrl`（http://localhost:5173）。
- `bun run build` 按 `buildCommand` 先构建前端再内嵌打包。
- `bun run typecheck` 类型检查；`bun run build:web` 只构建前端（调试产物用）。

## reader 页结构（src/reader/main.ts）

- 局部状态只保存当前文章、事件 revision 和影响渲染的偏好；没有路由或全局 store。
- 启动时订阅一次 `events.appEvent`；唯一 dispatcher 按 `eid` 路由，丢弃 `revision` 不
  递增的事件。消费的分支：`state.snapshot`/`appearance.changed`/`prefs.changed`（主题、
  字号、阅读字体/字号/行高与减少动效）、`selection.changed`（清空文章栈）、
  `post.selected`（打开正文）、`post.fulltext`（原生头部「提取全文」的按需结果，匹配
  当前文章时替换正文）。
- **文章层栈**：每篇文章一个绝对定位层，被盖住的层保留 DOM 和滚动位置；换篇动画由
  原生列表选择方向决定（`direction`：1 新文章从下方淡入、旧层上滑退出；-1 新层从上方
  盖下）。动画只在元素挂载后的下一帧启动（先内联钉住首帧），滑动用 Web Animations API，
  被替换的层有 `finished` 回调 + 超时兜底回收；减少动效偏好（`prefs.reduceMotion`）或系
  统 `prefers-reduced-motion` 下跳过滑动直接进终态（清理仍走 `onFinished`），正文翻页的
  平滑滚动同步降级。
- 文章请求用递增 requestId 防止快速切换时旧响应覆盖新文章；加载失败时恢复被盖住的
  旧层。
- **阅读进度条**：`#progress` 吸顶细条（`muted-foreground` 混色灰，随三套主题自适应）
  跟踪顶层文章层的滚动进度，可见更新经 CSS transition 缓动逼近，换篇重现时瞬时到位，
  减少动效时全程无动画；scroll 事件不冒泡，用 `#article` 上一处 capture 监听覆盖所有层；
  换篇、正文加载、`post.fulltext` 替换、字体/行高变化和窗口 resize 后重算；一屏读完的
  短文章隐藏。
- 打开正文由 `Posts.get` 完成；已读标记在 Go 侧按偏好自动处理（见 core/README），页面
  不再消费 `post.updated`、`m` 键调 `Posts.toggleRead` 服务端翻转。
- `post.fulltext` 的正文已在 Go 侧清洗（与 `Post.Content` 同一管线），页面按 feedId/guid
  匹配顶层文章层后直接进 `innerHTML`；内容不落库，切走再切回会回到订阅源提供的正文。
- 主题两层：`action.UpdatePreferences` 设置原生外观；页面在 `<html>` 上切换 `dark` 和
  `theme-sepia` 类，字号经 `--app-text-scale` 缩放。`system` 主题监听
  `prefers-color-scheme`。`sepia` 的运行时类名必须是 `theme-sepia`。桌面强调色与文字
  缩放经 `mygo-runtime` 的 `preferences()` 写入 `--mygo-accent`/`--mygo-text-scale`。
- 需要展示日期时使用 `zh-CN` Intl；需要测试确定性时把 `now` 或时间值作为参数传入。

## 安全约定

- `Post.Content` 是 Go 端已清洗的 HTML，可直接进入正文容器；其他不可信文本用
  Solid `{}` 插值（自动转义）或 `textContent`。禁止
  `fetch`/`XMLHttpRequest`/`WebSocket`/`EventSource`。
- 正文里的链接点击先经 `externalHref()` 过滤只留 `http:`/`https:`，再交
  `Settings.openLink`；图片点击走 `Posts.openImage` 原生预览。
- 正文底部的附件条（`post.enclosures`）全部用 DOM API + `textContent` 构建，地址再过一次
  `externalHref()`。音频用原生 `<audio controls>` 播放，WebView 直连附件地址，与正文内嵌
  媒体同语义（不经 Go 代理）；「下载附件」走 `Posts.downloadEnclosure` Go 绑定，经安全
  下载管线后由原生保存对话框选路径，页面不发起任何下载请求。
- 样式使用 `styles.css` 的语义变量和系统主题类，不写死主题颜色；页面字符串用简体中文，
  TS 标识符和注释用英文。
