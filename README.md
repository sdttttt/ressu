# 🥑 Ressu

一个随处可见，普通，没有什么特点的传统 **RSS 阅读器**。

<img width="1229" height="823" alt="图片" src="https://github.com/user-attachments/assets/6df8c8af-92b2-4ee6-82ea-ddd085944aa4" />

## Feature

- 支持添加（RSS / Atom / JSON Feed）, 从 OPML 批量导入订阅源，或在设置里导出
- 无后台，无通知，无定时拉取。只有应用回到前台时，会自动触发抓取；最近 5 分钟已抓取过的
  订阅源会跳过，手动刷新不受限制
- 支持文章生成完整长图，方便和群友分享
- 跟随系统/浅色/深色主题
- 支持快捷键，按下`?`可查看
- 基于 Webview & GPU 混合渲染的显示效果

> 默认使用Go远程资源代理访问文章中的图片资源，如果发现图裂，可以尝试`设置 > 远程资源访问 > 切换到[Webview模式]`。

## Credit

- 桌面构建 [egoist/mygo](https://github.com/egoist/mygo)
- 前端 [solid.js](https://github.com/solidjs/solid)
- 存储 [etcd-io/bbolt](https://github.com/etcd-io/bbolt)
- 关于我在无意间被[codex](https://github.com/openai/codex)变成废柴这件事 

## LICENSE

See [LICENSE](LICENSE) (0BSD发布，爱干嘛干嘛).
