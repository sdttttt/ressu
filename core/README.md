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
| `logging`     | 可选的日志文件封装                                              |
| `textmetrics` | 正文字数统计                                                    |
| `limitreader` | 带体积上限的 `io.Reader`                                        |

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
