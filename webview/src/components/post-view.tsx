import { ExternalLink, MailOpen, Newspaper, Star } from "lucide-solid";
import { For, Show, createEffect, onCleanup, onMount } from "solid-js";
import { createStore, reconcile } from "solid-js/store";

import { EmptyState } from "~/components/empty-state";
import { IconButton } from "~/components/icon-button";
import { Badge } from "~/components/ui/badge";
import { Separator } from "~/components/ui/separator";
import { Skeleton } from "~/components/ui/skeleton";
import { formatDateTime } from "~/lib/format";
import { sanitizeHTML } from "~/lib/sanitize";
import { cn } from "~/lib/utils";
import type { Post } from "~/mygo";
import { app, currentFeed, openLink, setOpenPostRead, toggleStar } from "~/stores/app";

/** How long the incoming article takes to cover the previous one, in milliseconds. */
const COVER_MS = 300;

/** How an article layer entered the screen; "base" layers sit still underneath. */
type LayerKind = "base" | "enter" | "next" | "prev";

type Layer = {
  /** The covered article keeps rendering until the slide is done. */
  kind: LayerKind;
  post: Post;
};

/**
 * One snapshot per swap, written by the effect below. Holding plain copies —
 * never the store proxy — matters: `setState("post", ...)` merges into the
 * same store node in place, so a held proxy would silently turn into the next
 * article mid-animation.
 */
const [layers, setLayers] = createStore<{ items: Layer[] }>({ items: [] });

let dropTimer: ReturnType<typeof setTimeout> | undefined;

/** How the new article should cover the old one, based on list order. */
function coverDirection(from: Post, to: Post): "next" | "prev" {
  const indexAt = (guid: string) => app.posts.findIndex((post) => post.guid === guid);
  const fromIndex = indexAt(from.guid);
  const toIndex = indexAt(to.guid);
  if (fromIndex === -1 || toIndex === -1) return "next";
  return toIndex > fromIndex ? "next" : "prev";
}

createEffect(() => {
  const latest = app.post;
  const visible = layers.items;

  if (!latest) {
    if (visible.length > 0) {
      setLayers({ items: [] });
    }
    clearTimeout(dropTimer);
    return;
  }

  const current = visible[visible.length - 1];
  // Same article: refresh the top snapshot in place (star/read toggles) and
  // never restart the animation.
  if (current?.post.guid === latest.guid) {
    setLayers("items", visible.length - 1, "post", reconcile({ ...latest }, { key: "guid" }));
    return;
  }

  const kind: LayerKind = current ? coverDirection(current.post, latest) : "enter";
  const next: Layer[] = [{ kind, post: { ...latest } }];
  // The article being covered freezes as a plain "base" copy underneath and
  // is dropped by the timer once the slide has finished.
  if (current) next.unshift({ kind: "base", post: { ...current.post } });
  setLayers({ items: next });

  clearTimeout(dropTimer);
  dropTimer = setTimeout(() => {
    setLayers("items", (items) => items.filter((layer) => layer.kind !== "base"));
  }, COVER_MS);
});

onCleanup(() => clearTimeout(dropTimer));

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
        <div class="relative min-h-0 flex-1 overflow-hidden">
          {/* Keyed by guid: the covered article stays perfectly still
              underneath, and the incoming one slides over it like a sheet of
              paper, from above or below depending on which way the reader
              moved in the list. Dropping the base entry from the array only
              removes that layer; the incoming one is never rebuilt, so its
              entry animation cannot replay. */}
          <For each={layers.items}>
            {(layer) => <ArticleLayer layer={layer.kind} post={layer.post} />}
          </For>
        </div>
      </Show>
    </article>
  );
}

/**
 * Mount-time animation for an article layer, played through the Web
 * Animations API. The layer also carries CSS animation classes describing the
 * same motion, but the two must never run at once: WebKit composes concurrent
 * animations on the same property by letting the CSS one win, which cancels
 * the slide and makes the swap look instant (or blank mid-flight). So the
 * API play cancels the CSS animations on the element; the classes stay only
 * as documentation of the motion and a fallback if `animate` is missing.
 */
function playEntry(ref: HTMLDivElement, layer: LayerKind): void {
  if (layer === "base") return;
  if (typeof ref.animate !== "function") return; // non-browser environment (tests)
  // Kill any CSS animation before it can start: the ref callback runs before
  // Solid applies the classes, so cancelling via getAnimations() would race
  // and WebKit would let the CSS animation win the composite. An inline
  // `animation: none` beats the classes at any timing and does not touch
  // WAAPI animations.
  ref.style.animation = "none";
  const frames =
    layer === "enter"
      ? [
          { opacity: "0", transform: "translateY(24px)" },
          { opacity: "1", transform: "translateY(0)" },
        ]
      : layer === "next"
        ? [{ transform: "translateY(-100%)" }, { transform: "translateY(0)" }]
        : [{ transform: "translateY(100%)" }, { transform: "translateY(0)" }];
  ref.animate(frames, { duration: COVER_MS, easing: "ease-out" });
}

function ArticleLayer(props: { layer: LayerKind; post: Post }) {
  const bindLayer = (el: HTMLDivElement) => playEntry(el, props.layer);

  return (
    <div
      ref={bindLayer}
      data-layer={props.layer}
      class={cn(
        "flex min-h-0 flex-1 flex-col",
        props.layer === "base" && "pointer-events-none absolute inset-0 z-0",
        props.layer === "enter" &&
          "animate-in duration-300 ease-out fade-in slide-in-from-bottom-6",
        props.layer === "next" &&
          "absolute inset-0 z-10 animate-in bg-background duration-300 ease-out slide-in-from-top-full",
        props.layer === "prev" &&
          "absolute inset-0 z-10 animate-in bg-background duration-300 ease-out slide-in-from-bottom-full",
      )}
    >
      <header class="border-b px-6 py-4">
        <div class="flex items-start gap-3">
          <div class="min-w-0 flex-1">
            <h1 class="text-lg leading-snug font-semibold">{props.post.title}</h1>
            <div class="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-[0.6875rem] text-muted-foreground">
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
          </div>
        </div>
      </header>

      <div class="min-h-0 flex-1 overflow-y-auto px-6 py-5">
        <div
          class="post-body mx-auto max-w-3xl"
          // The HTML is cleaned by DOMPurify before it reaches the DOM.
          innerHTML={sanitizeHTML(props.post.content || props.post.summary, props.post.url)}
        />
      </div>
    </div>
  );
}
