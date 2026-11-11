# WIP

本次提交是快照预览原生化的阶段性收尾，功能主体已完成，仍建议在真实桌面环境补一次人工回归。

## 已完成

- 原生文章头部新增“生成快照”，通过 `ui.command: "snapshot"` 通知静态阅读页。
- 静态阅读页保留离屏卡片、分片截图和进度提示，不再展示 WebView 内的预览与保存/复制操作。
- `Posts.StitchSnapshot` 拼接成功后由 Go 打开原生快照预览窗口。
- 原生预览窗口提供“保存”和“复制”，操作结果走原生通知；窗口保持打开。
- 已补充动作层和原生预览窗口测试，并同步 `AGENTS.md`。

## 待确认

- 在 Windows WebView2 下完整走一遍长文章、图片较多文章和窄窗口场景。
- 验证高 DPI 下拼接结果、保存对话框和剪贴板图片尺寸。
- 验证切换文章或关闭主窗口时，快照生成能正确取消并释放临时状态。

## 验证

- `go test ./internal/action ./internal/gui ./internal/webendpoint`
- `node --check webview/static/reader.js`
