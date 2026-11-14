# 🥑 Ressu

一个随处可见，普通，没有什么特点的传统 **RSS 阅读器**。

<img width="1363" height="842" alt="图片" src="https://github.com/user-attachments/assets/fb0e7561-fa50-4ba3-b1f2-5f87351f7d32" />

## Feature

- 支持添加（RSS / Atom / JSON Feed）, 从 OPML 批量导入订阅源
- 无后台，无通知，无定时拉取，应用回前台时，会自动触发抓取
- 可以对文章进行快照，方便与群友分享
- 跟随系统/浅色/深色主题
- 支持快捷键，按下`?`可查看
- 基于 GPU & Webview 的混合渲染，只有文章本身为Webview渲染，其他视觉组件均采用了MyGO的NativeUI

> 默认使用 WebView 直接加载文章中的图片资源，如果发现图裂，可以尝试`设置 > 远程资源访问 > 切换到[Go代理模式]`。

## Credit

- [egoist/mygo](https://github.com/egoist/mygo)
- [etcd-io/bbolt](https://github.com/etcd-io/bbolt)
- ~~关于我在无意间被[codex](https://github.com/openai/codex)变成废柴这件事~~

## LICENSE

See [LICENSE](LICENSE) (0BSD发布，爱干嘛干嘛).
