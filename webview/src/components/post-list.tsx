import {
  ArrowDownWideNarrow,
  ArrowUpNarrowWide,
  CheckCheck,
  FileText,
  Loader2,
  Star,
} from "lucide-solid";
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
  setPostSortOrder,
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
      <Show when={currentFeed()?.title ?? "所有"}>
        {(aggregateTitle) => (
          <>
            <header class="border-b px-3 py-2">
              <div class="truncate text-sm font-semibold">{aggregateTitle()}</div>
              <div class="mt-1 flex items-center gap-1">
                <div class="flex min-w-0 flex-1 items-center gap-1.5 text-[0.6875rem] text-muted-foreground">
                  <span class="tabular-nums">{app.postTotal} 篇</span>
                  <Separator orientation="vertical" class="h-3" />
                  <Badge
                    variant="secondary"
                    class="h-4 rounded-full px-1.5 py-0 text-[0.625rem] font-medium tabular-nums"
                  >
                    {app.postUnread} 未读
                  </Badge>
                </div>
                <IconButton
                  label={app.postSortOrder === "asc" ? "按时间降序排列" : "按时间升序排列"}
                  class="size-7"
                  onClick={() =>
                    void setPostSortOrder(app.postSortOrder === "asc" ? "desc" : "asc")
                  }
                >
                  <Show
                    when={app.postSortOrder === "asc"}
                    fallback={<ArrowDownWideNarrow class="size-3.5" />}
                  >
                    <ArrowUpNarrowWide class="size-3.5" />
                  </Show>
                </IconButton>
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
                  <ToggleGroupItem value="unread" size="sm" class="h-7 px-2 text-[0.6875rem]">
                    未读
                  </ToggleGroupItem>
                  <ToggleGroupItem value="starred" size="sm" class="h-7 px-2 text-[0.6875rem]">
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
                  <div class="space-y-1">
                    <For each={app.posts}>{(post) => <PostRow post={post} />}</For>
                  </div>

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
  const selected = () =>
    app.current.postFeedId === props.post.feedId && app.current.guid === props.post.guid;
  const open = () => void selectPost(props.post.guid, props.post.feedId);

  return (
    <div
      data-feed-id={props.post.feedId}
      data-guid={props.post.guid}
      class={cn(
        "group relative rounded-md transition-[background-color,transform] duration-100 ease-out hover:bg-accent/60 active:scale-[0.98]",
        selected() && "bg-accent",
      )}
    >
      <button type="button" class="w-full px-2.5 py-2 text-left" onClick={open}>
        <div
          class={cn(
            "line-clamp-2 text-[0.8125rem] leading-snug",
            props.post.read ? "text-muted-foreground" : "font-semibold",
          )}
        >
          {props.post.title}
        </div>

        <Show when={props.post.summary}>
          <p class="mt-1 line-clamp-3 text-[0.6875rem] leading-relaxed text-muted-foreground">
            {props.post.summary}
          </p>
        </Show>

        <div class="mt-1.5 flex items-center gap-2 pr-7 text-[0.625rem] text-muted-foreground tabular-nums">
          <Show when={!props.post.read}>
            <span class="size-1.5 rounded-full bg-primary" aria-hidden="true" />
          </Show>
          <span>{formatPostDate(props.post.publishedAt)}</span>
          <Show when={props.post.author}>
            <span class="max-w-32 truncate">{props.post.author}</span>
          </Show>
        </div>
      </button>

      <div class="absolute right-1 bottom-1">
        <IconButton
          label={props.post.starred ? "取消星标" : "加星标"}
          class={cn(
            "pointer-events-none size-6 opacity-0 transition-opacity group-hover:pointer-events-auto group-hover:opacity-100 focus-visible:pointer-events-auto focus-visible:opacity-100",
            props.post.starred && "pointer-events-auto text-warning opacity-100",
          )}
          onClick={() => void toggleStar(props.post.guid, props.post.feedId)}
        >
          <Star class={cn("size-3.5", props.post.starred && "fill-current")} />
        </IconButton>
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
