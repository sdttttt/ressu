import { CheckCheck, FileText, Loader2, Star } from "lucide-solid";
import { For, Show } from "solid-js";

import { EmptyState } from "~/components/empty-state";
import { IconButton } from "~/components/icon-button";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Separator } from "~/components/ui/separator";
import { Skeleton } from "~/components/ui/skeleton";
import { ToggleGroup, ToggleGroupItem } from "~/components/ui/toggle-group";
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

function activeFilters(): string[] {
  const filters: string[] = [];
  if (app.unreadOnly) filters.push("unread");
  if (app.starredOnly) filters.push("starred");
  return filters;
}

export default function PostList() {
  return (
    <section class="flex h-full min-h-0 flex-col bg-background">
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
                  <div class="mt-1 flex items-center gap-1.5 text-[11px] text-muted-foreground">
                    <span class="tabular-nums">{app.postTotal} 篇</span>
                    <Separator orientation="vertical" class="h-3" />
                    <Badge
                      variant="secondary"
                      class="h-4 rounded-full px-1.5 py-0 text-[10px] font-medium tabular-nums"
                    >
                      {app.postUnread} 未读
                    </Badge>
                  </div>
                </div>

                <ToggleGroup
                  multiple
                  value={activeFilters()}
                  onChange={(values) =>
                    void setPostFilter({
                      unreadOnly: values.includes("unread"),
                      starredOnly: values.includes("starred"),
                    })
                  }
                  aria-label="筛选文章"
                >
                  <ToggleGroupItem value="unread" size="sm" class="h-7 px-2 text-[11px]">
                    未读
                  </ToggleGroupItem>
                  <ToggleGroupItem value="starred" size="sm" class="h-7 px-2 text-[11px]">
                    星标
                  </ToggleGroupItem>
                </ToggleGroup>

                <IconButton
                  label="全部标记为已读"
                  class="size-7"
                  onClick={() => void markAllRead()}
                >
                  <CheckCheck class="size-3.5" />
                </IconButton>
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
          <IconButton
            label={props.post.starred ? "取消星标" : "加星标"}
            class={cn(
              "size-6 opacity-0 transition-opacity group-hover:opacity-100 focus-visible:opacity-100",
              props.post.starred && "text-warning opacity-100",
            )}
            onClick={() => void toggleStar(props.post.guid)}
          >
            <Star class={cn("size-3.5", props.post.starred && "fill-current")} />
          </IconButton>
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
