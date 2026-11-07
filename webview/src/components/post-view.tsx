import { Camera, ExternalLink, Loader2, MailOpen, Newspaper, Star } from "lucide-solid";
import { For, Show, createEffect, onCleanup, onMount } from "solid-js";
import { createStore, reconcile } from "solid-js/store";

import ArticleLayer, { type LayerKind } from "~/components/article-layer";
import { EmptyState } from "~/components/empty-state";
import { IconButton } from "~/components/icon-button";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Separator } from "~/components/ui/separator";
import { Skeleton } from "~/components/ui/skeleton";
import { formatDateTime } from "~/lib/format";
import { sanitizeHTML } from "~/lib/sanitize";
import { cn } from "~/lib/utils";
import type { Post } from "~/mygo";
import {
  app,
  captureSnapshot,
  currentFeed,
  openLink,
  setOpenPostRead,
  toggleStar,
} from "~/stores/app";

/** How long the article pane keeps a swap's leftover layers around, in
 * milliseconds: twice the slide plus a frame of slack, for engines that
 * never report the slide as finished. */
const SWAP_CLEANUP_MS = 700;

type Layer = {
  kind: LayerKind;
  post: Post;
  /** The swap this layer belongs to. */
  swap?: number;
  /** Set on leftover layers a swap drops once its slide has finished. */
  retire?: boolean;
};

/**
 * One snapshot per swap, written by the effect below. Holding plain copies —
 * never the store proxy — matters: `setState("post", ...)` merges into the
 * same store node in place, so a held proxy would silently turn into the next
 * article mid-animation.
 */
const [layers, setLayers] = createStore<{ items: Layer[] }>({ items: [] });

let cleanupTimer: ReturnType<typeof setTimeout> | undefined;
let swapId = 0;

/** How the swap should animate, based on list order. */
function swapDirection(from: Post, to: Post): "cover-down" | "exit-up" {
  const indexAt = (guid: string) => app.posts.findIndex((post) => post.guid === guid);
  const fromIndex = indexAt(from.guid);
  const toIndex = indexAt(to.guid);
  if (fromIndex === -1 || toIndex === -1) return "cover-down";
  // Moving earlier in the list: the previous article sheets down from the
  // top. Moving later: the current article slides up and away, revealing the
  // next one beneath it.
  return toIndex > fromIndex ? "exit-up" : "cover-down";
}

createEffect(() => {
  const latest = app.post;
  const visible = layers.items;

  if (!latest) {
    if (visible.length > 0) {
      setLayers({ items: [] });
    }
    clearTimeout(cleanupTimer);
    return;
  }

  const current = visible[visible.length - 1];
  // Same article: refresh the top snapshot in place (star/read toggles) and
  // never restart the animation.
  if (current?.post.guid === latest.guid) {
    setLayers("items", visible.length - 1, "post", reconcile({ ...latest }, { key: "guid" }));
    return;
  }

  const swap = ++swapId;
  // The article being replaced keeps its object: the same reference makes
  // <For> reuse that layer's DOM node, so images and scroll position survive
  // untouched. Its new kind decides the choreography: cover-down freezes it
  // as the backdrop the previous article sheets over; exit-up turns it into
  // the layer that slides away.
  const covered = current ?? undefined;
  const direction = covered ? swapDirection(covered.post, latest) : "enter";
  const entering: Layer = {
    kind: direction === "cover-down" ? "cover-down" : "enter",
    post: { ...latest },
    swap,
  };
  if (covered) {
    setLayers("items", visible.length - 1, {
      kind: direction === "exit-up" ? "exit-up" : "base",
      swap,
      retire: true,
    });
  }
  setLayers("items", covered ? [visible[visible.length - 1], entering] : [entering]);

  clearTimeout(cleanupTimer);
  cleanupTimer = setTimeout(() => dropRetired(swap), SWAP_CLEANUP_MS);
});

/** Drops the leftover layer of one swap once its slide has finished. */
function dropRetired(swap: number): void {
  setLayers("items", (items) => items.filter((layer) => !(layer.retire && layer.swap === swap)));
}

onCleanup(() => clearTimeout(cleanupTimer));

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
        when={layers.items.length > 0}
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
            <div class="w-full space-y-3 px-6 py-6">
              <Skeleton class="h-6 w-2/3" />
              <Skeleton class="h-3 w-40" />
              <Skeleton class="mt-6 h-3 w-full" />
              <Skeleton class="h-3 w-11/12" />
              <Skeleton class="h-3 w-10/12" />
            </div>
          </Show>
        }
      >
        <div class="relative min-h-0 flex-1 overflow-hidden">
          {/* Keyed by layer object: a replaced article keeps its node (just
              flipping kind), so its state never resets and the incoming
              slide cannot replay on it. */}
          <For each={layers.items}>
            {(layer) => (
              <ArticleLayer kind={layer.kind} swap={layer.swap} onFinished={dropRetired}>
                <ArticleBody post={layer.post} />
              </ArticleLayer>
            )}
          </For>
        </div>
      </Show>
    </article>
  );
}

/** One article's content: header row plus the sanitized body. */
function ArticleBody(props: { post: Post }) {
  return (
    <>
      <header class="border-b px-6 py-4">
        <h1 class="text-lg leading-snug font-semibold">{props.post.title}</h1>
        <div class="mt-1.5 flex items-center gap-3">
          <div class="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 text-[0.6875rem] text-muted-foreground">
            <span class="tabular-nums">{formatDateTime(props.post.publishedAt)}</span>
            <Show when={props.post.author}>
              {(author) => (
                <>
                  <Separator orientation="vertical" class="h-3" />
                  <span>{author()}</span>
                </>
              )}
            </Show>
            <Show when={currentFeed()}>
              {(feed) => (
                <>
                  <Separator orientation="vertical" class="h-3" />
                  <Badge
                    variant="outline"
                    class="h-4 max-w-48 rounded-full px-1.5 py-0 text-[0.625rem] font-normal"
                  >
                    <span class="truncate">{feed().title}</span>
                  </Badge>
                </>
              )}
            </Show>
          </div>
          <div class="flex shrink-0 items-center gap-0.5">
            <IconButton
              label={props.post.starred ? "取消星标" : "加星标"}
              onClick={() => void toggleStar(props.post.guid)}
            >
              <Star class={cn("size-4", props.post.starred && "fill-current text-warning")} />
            </IconButton>
            <IconButton
              label={props.post.read ? "标记为未读" : "标记为已读"}
              onClick={() => void setOpenPostRead(!props.post.read)}
            >
              <MailOpen class="size-4" />
            </IconButton>
            <IconButton
              label="在浏览器中打开原文"
              disabled={!props.post.url}
              onClick={() => void openLink(props.post.url)}
            >
              <ExternalLink class="size-4" />
            </IconButton>
            <Button
              variant="ghost"
              size="icon"
              class="size-8 shrink-0"
              title="生成文章快照"
              aria-label="生成文章快照"
              disabled={
                app.snapshotting ||
                app.snapshotAction !== null ||
                (!props.post.content && !props.post.summary)
              }
              onClick={() => void captureSnapshot()}
            >
              <Show when={app.snapshotting} fallback={<Camera class="size-4" />}>
                <Loader2 class="size-4 animate-spin" />
              </Show>
            </Button>
          </div>
        </div>
      </header>

      <div data-post-scroll class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
        <div
          class="post-body"
          // The HTML is cleaned by DOMPurify before it reaches the DOM.
          innerHTML={sanitizeHTML(props.post.content || props.post.summary, props.post.url)}
        />
      </div>
    </>
  );
}
