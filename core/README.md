# Ressu Core

Ressu 阅读器的后端核心。

该包独立仅仅是为了和GUI解耦，暂时没有移植核心的打算。

## 包一览

| 包            | 作用                                                           |
| ------------- | -------------------------------------------------------------- |
| `rss`         | 订阅源抓取与解析：RSS 0.9x/1.0/2.0、Atom、JSON Feed、OPML 读写  |
| `sanitize`    | 基于 bluemonday 的正文 HTML 清洗与链接绝对化                   |
| `icon`        | 订阅源图标发现、下载校验、体积限制                              |
| `store`       | bbolt 单文件持久化：订阅、文章、偏好、图标缓存、迁移            |
| `media`       | 正文图片安全下载、磁盘缓存；缓存本身是一个 `http.Handler`       |
| `service`     | 业务入口：订阅增删刷新、文章列表/搜索/分页、偏好、快照拼接      |
| `event`       | 统一事件 DTO 与串行发布器 Hub（sink 注入，发布者不带传输层）    |
| `status`      | 服务状态的内存镜像模型（供状态栏等只读消费）                    |
| `logging`     | 可选的日志文件封装；`Debug` 仅在 `debug` build tag 下输出        |
| `textmetrics` | 正文字数统计                                                    |
| `limitreader` | 带体积上限的 `io.Reader`                                        |

`service` 不 import `mygo` 或任何 GUI 包；`event` 的 sink 注入设计保证发布器不依赖传输层。

## 订阅源抓取与解析

- `rss.ValidateURL` 只接受带 host 的绝对 `http`/`https` URL；`file:`、`javascript:` 和相对
  地址都返回 `ErrInvalidURL`。
- `rss.Fetch` 发送 `User-Agent: Ressu/0.1 (+https://github.com/sdttttt/ressu)`、偏向订阅源的
  `Accept`，以及可用的 `If-None-Match`/`If-Modified-Since`。请求和读取 body 共享 15 秒超时；
  响应最多读取 16 MiB；非 200/304 状态直接报错。
- `rss.Parse` 使用 gofeed 读取 RSS 0.9x/1.0/2.0、Atom 和 JSON Feed。GUID 缺失时先退回
  链接，再退回“标题 + 发布时间”；标题也缺失时使用 GUID。频道和文章里的相对链接按订阅源
  地址解析为绝对地址，摘要去掉 HTML、折叠空白并截断。
- 正文在 `core/sanitize` 基于 bluemonday 清洗后入库；链接和媒体地址按文章地址补全。
  损坏 XML 或不是订阅源的页面返回错误，由服务层展示给用户。
- 订阅源图标由 `core/icon` 在 Go 端解析：优先 feed 内置图标，其次站点 favicon 链接，最后
  `/favicon.ico`。只允许位图、http/https、256KB 以内，并通过 bbolt 缓存。

## 刷新流程与事件

- `Feeds.Add` 先校验 URL、检查重复，再立即抓取一次并保存频道元数据与首批文章；成功后
  广播 `feeds.changed` 和 `posts.changed`。`Feeds.Update` 修改订阅源地址或名称：地址变更会
  立即重抓并合并文章（重置 ETag/Last-Modified），查重排除自身；非空名称写入
  `Feed.CustomTitle`，`refreshOne` 刷新元数据时跳过自定义标题，清空名称则下次刷新回落到
  频道标题。`Feeds.Test` 校验并预览地址，查重同样支持排除自身（编辑时复测原地址）。
- `Feeds.RefreshAll` 用手动全量刷新，`maxConcurrent = 2` 信号量并发，每个订阅完成广播
  `sync.progress`，批次开始/结束广播 `sync.started`/`sync.finished`。
- `Feeds.RefreshStale` 供启动和回前台自动刷新：只抓 `LastFetchedAt` 为空或早于 5 分钟的
  订阅源，复用全量刷新的单飞与并发逻辑；已有刷新时返回 `ErrRefreshing`。
- `refreshOne` 把 ETag 与 Last-Modified 交给 `rss.Fetch`；304 映射为 `rss.ErrNotModified`
  （保留文章，只清除错误并更新 `lastFetchedAt`）。失败把错误写入 `Feed.LastError` 并广播
  `feeds.changed`，订阅和已存文章保留。
- 成功抓取后调用 `store.MergePosts`，再刷新频道元数据与校验器；合并后重新读取 Feed，
  因为未读数已在事务中重算。
- OPML：`core/rss/opml.go` 负责解析（跳过无 `xmlUrl` 的目录）和导出；
  `core/service/opml.go` 汇总导入结果（`ImportSummary`），导入复用 `Feeds.Add`。
- `Posts.List` 在服务端完成未读/星标筛选、搜索和分页，返回 `{posts,total,unread}`；
  默认每页 100、上限 500。未读数由文章推导并保存在 `Feed.Unread`：合并时重算，单条已读
  变化时增减，全部已读时清零。
- `Posts.Get` 按偏好 `MarkReadOnOpen` 自动标已读（失败不阻塞阅读）并广播
  `post.updated`；`Posts.ToggleRead` 翻转已读并返回最新摘要。
- 事件由 `core/event.Hub` 串行发布：先分配递增 `revision`、更新时间，再更新 `status.Model`
  镜像，最后交给 sink。任何改变订阅或未读状态的服务方法都应广播。

## 持久化

数据库是单文件 bbolt（`ressu.db`，文件模式 0600）：

| 桶         | 键                | 值                                                      |
| ---------- | ----------------- | ------------------------------------------------------- |
| `feeds`    | 8 字节大端订阅 id | 整个 `Feed` 的 JSON：元数据、ETag/Last-Modified、未读数 |
| `posts`    | 订阅 id（嵌套桶） | 每个 GUID 一条 `Post` 的 JSON                           |
| `settings` | `prefs`           | 整个 `Prefs` 的 JSON                                    |
| `icons`    | 订阅 id（嵌套桶） | `data` 键存原始图片字节；`meta` 键存缓存元数据 JSON     |

- `Store` 在内存缓存订阅和偏好，文章每次从数据库读取；一把互斥锁串行化写操作；返回的
  slice 会克隆。bbolt 的文件锁意味着同一数据目录只能有一个进程打开数据库。
- `settings/schemaVersion` 记录数据库版本；修改持久化结构必须兼容旧记录，或在
  `store/migrate.go` 加入迁移。
- `MergePosts` 按 GUID 合并：已有文章保留 `read`、`starred`、`fetchedAt`，其余字段刷新；
  按 `publishedAt` 倒序，裁剪到 `Prefs.PostsPerFeed`，并重算未读；它在一个事务里重写该
  订阅的整个文章桶。`UpdatePost` 和 `MarkAllRead` 只写受影响的数据。
- `AddFeed` 使用“当前最大 id + 1”；删除最高 id 后该 id 会被复用。
- 远程资源访问方式全局默认存于 `Prefs.RemoteResources`；每个订阅源可用
  `Feed.RemoteResources` 覆盖（空串 = 跟随全局，新增时由界面写入当时的全局值），
  `Posts.Get` 按文章所属订阅源的判断决定是否经 `media.RewriteHTML` 走本地缓存。
- `Prefs.Normalize` 替换越界值：未知主题回落 `system`，未知字号回落 `medium`，
  `postSort` 不是 `fetched` 就回落 `published`，`openLinks` 不是 `window` 就变成
  `browser`，`remoteResources` 不是 `webview` 就回落 `goProxy`，`postsPerFeed` 限制在
  50–2000（默认 500），`imageCacheLimit` 限制在 32–512 MB（默认 256），面板宽度恢复
  `[18,30]`。偏好默认值只以 `store.DefaultPrefs` 为准。
- `Post.Content` 是入库前清洗后的 HTML，`Post.Summary` 是截到 280 个字符的纯文本。

## 示例代码

```go
package main

import (
	"context"
	"log"

	"github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/core/service"
	"github.com/sdttttt/ressu/core/status"
	"github.com/sdttttt/ressu/core/store"
)

func main() {
	// 1. 打开（或创建）数据库：用户数据目录下的 ressu.db。
	st, err := store.Open("./data")
	if err != nil {
		log.Fatal(err)
	}
	defer st.Close()

	// 2. 创建事件 Hub。sink 是事件出口：推给 SSE、WebSocket、
	//    终端日志或任何你的前端；传 nil 则只更新内存镜像。
	model := status.New("0.1.0")
	hub := event.NewHub(model, func(e event.Event) error {
		log.Printf("event %s rev=%d", e.EID, e.Revision)
		return nil
	})

	// 3. 组合业务服务。
	feeds := service.NewFeedsWithHub(st, hub)
	posts := service.NewPostsWithMediaAndHub(st, nil, hub)
	settings := service.NewSettings(st)

	// 4. 正常使用：订阅、刷新、读文章。
	ctx := context.Background()
	feed, err := feeds.Add(ctx, "", "https://example.com/feed.xml")
	if err != nil {
		log.Fatal(err)
	}
	summary, err := feeds.RefreshAll(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("updated=%d new=%d failed=%d", summary.Updated, summary.New, summary.Failed)

	page, err := posts.List(service.PostQuery{FeedID: feed.ID})
	if err != nil {
		log.Fatal(err)
	}
	for _, post := range page.Posts {
		log.Println(post.Title)
	}
	_ = settings
}
```
