import { CheckCheck, FileText, Loader2, Star } from "lucide-solid";
import { For, Show } from "solid-js";

import { EmptyState } from "~/components/empty-state";
import { Button } from "~/components/ui/button";
import { Skeleton } from "~/components/ui/skeleton";
import { formatPostDate } from "~/lib/format";
import { cn } from "~/lib/utils";
import type { PostSummary } from "~/mygo";
import {
  app,
  currentFeed,
  loadMorePosts,
  markAllRead,
  selectPost,
  setPostFilter,
  toggleStar,
} from "~/stores/app";

export default function PostList() {
  return (
    <section class="flex h-full min-h-0 flex-col border-r bg-background">
      <Show
        when={currentFeed()}
        fallback={<EmptyState icon={<FileText class="size-10" />} title="选择一个订阅源" />}
      >
        {(feed) => (
          <>
            <header class="border-b px-3 py-2">
              <div class="flex items-center gap-1">
                <div class="min-w-0 flex-1">
                  <div class="truncate text-sm font-semibold">{feed().title}</div>
                  <div class="text-[11px] text-muted-foreground tabular-nums">
                    {app.postTotal} 篇 · {app.postUnread} 未读
                  </div>
                </div>
                <Button
                  variant={app.unreadOnly ? "secondary" : "ghost"}
                  size="sm"
                  class="h-7 px-2 text-[11px]"
                  onClick={() => void setPostFilter({ unreadOnly: !app.unreadOnly })}
                >
                  未读
                </Button>
                <Button
                  variant={app.starredOnly ? "secondary" : "ghost"}
                  size="sm"
                  class="h-7 px-2 text-[11px]"
                  onClick={() => void setPostFilter({ starredOnly: !app.starredOnly })}
                >
                  星标
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  class="size-7"
                  title="全部标记为已读"
                  onClick={() => void markAllRead()}
                >
                  <CheckCheck class="size-3.5" />
                </Button>
              </div>
            </header>

            <div class="min-h-0 flex-1 overflow-y-auto px-2 py-2">
              <Show when={!app.loadingPosts || app.posts.length > 0} fallback={<PostSkeletons />}>
                <Show
                  when={app.posts.length > 0}
                  fallback={
                    <EmptyState
                      icon={<FileText class="size-9" />}
                      title={
                        app.unreadOnly || app.starredOnly ? "没有符合条件的文章" : "暂时没有文章"
                      }
                      hint={
                        app.unreadOnly || app.starredOnly
                          ? "关掉筛选，或者刷新这个订阅源。"
                          : "刷新订阅源后再来看看。"
                      }
                    />
                  }
                >
                  <For each={app.posts}>{(post) => <PostRow post={post} />}</For>

                  <Show when={app.posts.length < app.postTotal}>
                    <Button
                      variant="ghost"
                      size="sm"
                      class="mt-2 w-full text-xs text-muted-foreground"
                      disabled={app.loadingPosts}
                      onClick={() => void loadMorePosts()}
                    >
                      <Show
                        when={app.loadingPosts}
                        fallback={`加载更多（还有 ${app.postTotal - app.posts.length} 篇）`}
                      >
                        <Loader2 class="size-3.5 animate-spin" />
                        加载中…
                      </Show>
                    </Button>
                  </Show>
                </Show>
              </Show>
            </div>
          </>
        )}
      </Show>
    </section>
  );
}

function PostRow(props: { post: PostSummary }) {
  const selected = () => app.current.guid === props.post.guid;
  const open = () => void selectPost(props.post.guid);

  return (
    <div
      class={cn(
        "group relative rounded-md transition-colors hover:bg-accent/60",
        selected() && "bg-accent",
      )}
    >
      <div class="flex items-start gap-2 px-2.5 py-2">
        <button type="button" class="min-w-0 flex-1 text-left" onClick={open}>
          <div
            class={cn(
              "line-clamp-2 text-[13px] leading-snug",
              props.post.read ? "text-muted-foreground" : "font-semibold",
            )}
          >
            {props.post.title}
          </div>

          <Show when={props.post.summary}>
            <p class="mt-1 line-clamp-3 text-[11px] leading-relaxed text-muted-foreground">
              {props.post.summary}
            </p>
          </Show>

          <div class="mt-1.5 flex items-center gap-2 text-[10px] text-muted-foreground tabular-nums">
            <span>{formatPostDate(props.post.publishedAt)}</span>
            <Show when={props.post.author}>
              <span class="max-w-32 truncate">{props.post.author}</span>
            </Show>
          </div>
        </button>

        <div class="flex shrink-0 flex-col items-center gap-1.5">
          <Show when={!props.post.read}>
            <span class="size-1.5 rounded-full bg-primary" />
          </Show>
          <button
            type="button"
            class={cn(
              "rounded p-0.5 opacity-0 transition-opacity group-hover:opacity-100 hover:text-foreground focus-visible:opacity-100",
              props.post.starred ? "text-warning opacity-100" : "text-muted-foreground",
            )}
            title={props.post.starred ? "取消星标" : "加星标"}
            onClick={() => void toggleStar(props.post.guid)}
          >
            <Star class={cn("size-3.5", props.post.starred && "fill-current")} />
          </button>
        </div>
      </div>
    </div>
  );
}

function PostSkeletons() {
  return (
    <div class="space-y-2">
      <For each={[0, 1, 2, 3, 4]}>
        {() => (
          <div class="space-y-2 rounded-md px-2.5 py-2">
            <Skeleton class="h-3.5 w-3/4" />
            <Skeleton class="h-2.5 w-full" />
            <Skeleton class="h-2.5 w-5/6" />
          </div>
        )}
      </For>
    </div>
  );
}
