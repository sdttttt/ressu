# Ressu 应用门面（core/app）

`App` 是所有页面和原生界面能力的**唯一应用入口**（入站端口）：持有 service、事件 Hub 和
`UI` 回调接口，负责文件对话框、剪贴板、外链、截图、原生主题与状态广播。分层规则：

- 需要暴露给界面层的能力必须先定义成 `App` 方法；service 层不感知 UI。
- `app` 不 import `mygo`/`mygoui`：`UI` 是接口，只使用视图无关类型——
  `WindowHandle`（不透明窗口句柄，`NewWindowHandle` 包装具体窗口）、`Notification`、
  `ConfirmationRequest`、`AddFeedCallbacks`、`SettingsCallbacks` 等 DTO。桌面壳的
  `mygoui.NativeUI` 负责解包句柄并适配各原生窗口的回调结构体。
- 只有 `App` 负责事件广播和 UI 副作用；`mygoui` 组件只通过回调或只读数据
  类型接入，不直接访问数据库或实现业务规则。

## 与 webendpoint 的分工

`internal/webendpoint` 只做 MyGO 绑定适配：持有 `*app.App` 和唯一的广播通道
`AppEvent`（`mygo.NewEvent[appevent.Event]("app:event")`，以 sink 形式注入
`event.Hub`），导出方法里用 `mygo.CallerWindow(ctx)`/`mygo.CallerPage(ctx)` 取出调用方
窗口后转交 action。它不得直接持有或调用 service，不得复制业务逻辑，也不得自行拼装 UI
副作用。绑定对象共五个：`Feeds`、`Posts`、`Settings`、`Snapshot`、`Status`；导出方法
首参 `context.Context` 由 MyGO 注入，参数和返回值必须可 JSON 序列化（可空时间用
`*time.Time`，页面收到 `string | null`）。
