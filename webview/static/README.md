# Ressu 阅读页前端（webview/static）

无框架、无打包器的静态 HTML/CSS/JS，开发时由 MyGO 直供、打包时内嵌。页面**没有网络
权限**：所有数据走 `globalThis.mygo.call`，事件来自唯一的 `app:event` 通道。

## 文件职责

| 文件           | 职责                                                           |
| -------------- | -------------------------------------------------------------- |
| `index.html`   | 阅读页入口，`lang="zh-CN"`                                     |
| `reader.css`   | 正文排版、主题 token、空态和文章滑动动画样式                   |
| `reader.js`    | 阅读页唯一脚本：绑定调用、事件分发、文章层栈动画、阅读快捷键   |
| `snapshot.html` | 快照采集页入口，独立子窗口加载                                |
| `snapshot.css` | 快照卡片固定浅色样式和采集进度样式                             |
| `snapshot.js`  | 快照卡片重建、滚动分片采集和拼接调用                           |

## reader.js 结构

- 局部状态只保存当前文章、事件 revision 和影响渲染的偏好；没有路由或全局 store。
- 启动时订阅一次 `app:event`；唯一 dispatcher 按 `eid` 路由，丢弃 `revision` 不递增的
  事件。消费的分支：`state.snapshot`/`appearance.changed`/`prefs.changed`（主题与字号）、
  `selection.changed`（清空文章栈）、`post.selected`（打开正文）。
- **文章层栈**：每篇文章一个绝对定位层，被盖住的层保留 DOM 和滚动位置；换篇动画由
  原生列表选择方向决定（`direction`：1 新文章从下方淡入、旧层上滑退出；-1 新层从上方
  盖下）。动画只在元素挂载后的下一帧启动（先内联钉住首帧），滑动用 Web Animations API，
  被替换的层有 `finished` 回调 + 超时兜底回收。
- 文章请求用递增 requestId 防止快速切换时旧响应覆盖新文章；加载失败时恢复被盖住的
  旧层。
- 打开正文由 `Posts.Get` 完成；已读标记在 Go 侧按偏好自动处理（见 core/README），页面
  不再消费 `post.updated`、`m` 键调 `Posts.ToggleRead` 服务端翻转。
- 主题两层：`action.UpdatePreferences` 设置原生外观；页面在 `<html>` 上切换 `dark` 和
  `theme-sepia` 类，字号经 `--app-text-scale` 缩放。`system` 主题监听
  `prefers-color-scheme`。`sepia` 的运行时类名必须是 `theme-sepia`。
- 需要展示日期时使用 `zh-CN` Intl；需要测试确定性时把 `now` 或时间值作为参数传入。

## 安全约定

- `Post.Content` 是 Go 端已清洗的 HTML，可直接进入正文容器；其他不可信文本用
  `textContent` 写入。禁止 `fetch`/`XMLHttpRequest`/`WebSocket`/`EventSource`。
- 正文里的链接点击先经 `externalHref()` 过滤只留 `http:`/`https:`，再交
  `Settings.OpenLink`；图片点击走 `Posts.OpenImage` 原生预览。
- 样式使用 `reader.css` 的语义变量和系统主题类，不写死主题颜色；页面字符串用简体中文，
  JS 标识符和注释用英文。
