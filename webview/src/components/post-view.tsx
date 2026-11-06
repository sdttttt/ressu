import { ExternalLink, MailOpen, Newspaper, Star } from "lucide-solid";
import { Show, onCleanup, onMount } from "solid-js";

import { EmptyState } from "~/components/empty-state";
import { IconButton } from "~/components/icon-button";
import { Badge } from "~/components/ui/badge";
import { Separator } from "~/components/ui/separator";
import { Skeleton } from "~/components/ui/skeleton";
import { formatDateTime } from "~/lib/format";
import { sanitizeHTML } from "~/lib/sanitize";
import { cn } from "~/lib/utils";
import { app, currentFeed, openLink, setOpenPostRead, toggleStar } from "~/stores/app";

function handleLinkClick(event: MouseEvent, root: HTMLElement): void {
  const target = event.target as HTMLElement | null;
  const anchor = target?.closest("a");
  if (!anchor || !root.contains(anchor)) return;

  const href = anchor.getAttribute("href");
  if (!href) return;
  event.preventDefault();
  void openLink(href);
}

export default function PostView() {
  let articleRef: HTMLElement | undefined;

  const bindArticle = (element: HTMLElement) => {
    articleRef = element;
  };

  onMount(() => {
    const element = articleRef;
    if (!element) return;

    const listener = (event: MouseEvent) => handleLinkClick(event, element);
    element.addEventListener("click", listener);
    onCleanup(() => element.removeEventListener("click", listener));
  });

  return (
    <article ref={bindArticle} class="flex h-full min-h-0 flex-col bg-background">
      <Show
        when={app.post}
        fallback={
          <Show
            when={app.loadingPost}
            fallback={
              <EmptyState
                icon={<Newspaper class="size-12" />}
                title="选择一篇文章"
                hint="左侧列表里的文章会显示在这里。"
              />
            }
          >
            <div class="mx-auto w-full max-w-3xl space-y-3 px-6 py-6">
              <Skeleton class="h-6 w-2/3" />
              <Skeleton class="h-3 w-40" />
              <Skeleton class="mt-6 h-3 w-full" />
              <Skeleton class="h-3 w-11/12" />
              <Skeleton class="h-3 w-10/12" />
            </div>
          </Show>
        }
      >
        {(post) => (
          <>
            <header class="border-b px-6 py-4">
              <div class="flex items-start gap-3">
                <div class="min-w-0 flex-1">
                  <h1 class="text-lg leading-snug font-semibold">{post().title}</h1>
                  <div class="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-[11px] text-muted-foreground">
                    <span class="tabular-nums">{formatDateTime(post().publishedAt)}</span>
                    <Show when={post().author}>
                      <Separator orientation="vertical" class="h-3" />
                      <span>{post().author}</span>
                    </Show>
                    <Show when={currentFeed()}>
                      {(feed) => (
                        <>
                          <Separator orientation="vertical" class="h-3" />
                          <Badge
                            variant="outline"
                            class="h-4 max-w-48 rounded-full px-1.5 py-0 text-[10px] font-normal"
                          >
                            <span class="truncate">{feed().title}</span>
                          </Badge>
                        </>
                      )}
                    </Show>
                  </div>
                </div>

                <div class="flex shrink-0 items-center gap-0.5">
                  <IconButton
                    label={post().starred ? "取消星标" : "加星标"}
                    onClick={() => void toggleStar(post().guid)}
                  >
                    <Star class={cn("size-4", post().starred && "fill-current text-warning")} />
                  </IconButton>
                  <IconButton
                    label={post().read ? "标记为未读" : "标记为已读"}
                    onClick={() => void setOpenPostRead(!post().read)}
                  >
                    <MailOpen class="size-4" />
                  </IconButton>
                  <IconButton
                    label="在浏览器中打开原文"
                    disabled={!post().url}
                    onClick={() => void openLink(post().url)}
                  >
                    <ExternalLink class="size-4" />
                  </IconButton>
                </div>
              </div>
            </header>

            <div class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
              <div
                class="post-body mx-auto max-w-3xl"
                // The HTML is cleaned by DOMPurify before it reaches the DOM.
                innerHTML={sanitizeHTML(post().content || post().summary, post().url)}
              />
            </div>
          </>
        )}
      </Show>
    </article>
  );
}
