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

/** How an article layer behaves in the stack. */
type LayerKind = "base" | "enter" | "cover-down" | "exit-up";

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

let dropTimer: ReturnType<typeof setTimeout> | undefined;
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

  const swap = ++swapId;
  // The article being replaced keeps its DOM node: the same reference makes
  // <For> reuse it, so images and scroll position survive untouched. Its new
  // kind decides the choreography: cover-down freezes it as the backdrop the
  // previous article sheets over; exit-up turns it into the layer that
  // slides away (that kind change also triggers its animation below).
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
  const stack: Layer[] = covered ? [visible[visible.length - 1], entering] : [entering];
  setLayers("items", stack);

  clearTimeout(dropTimer);
  dropTimer = setTimeout(() => dropFinished(swap), COVER_MS * 2 + 100);
});

/** Drops the leftover layer of one swap once its slide has finished. */
function dropFinished(swap: number): void {
  setLayers("items", (items) => items.filter((layer) => !(layer.retire && layer.swap === swap)));
}

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
          <For each={layers.items}>{(layer) => <ArticleLayer layer={layer} />}</For>
        </div>
      </Show>
    </article>
  );
}

/**
 * Layer animation keyframes, keyed by kind. Entering rises 24px while
 * fading in; cover-down sheets down over the old article from above; the
 * exit-up layer slides up and away, revealing the incoming one beneath.
 */
const KIND_FRAMES: Record<Exclude<LayerKind, "base">, Keyframe[]> = {
  enter: [
    { opacity: "0", transform: "translateY(24px)" },
    { opacity: "1", transform: "translateY(0)" },
  ],
  "cover-down": [
    { opacity: "1", transform: "translateY(-100%)" },
    { opacity: "1", transform: "translateY(0)" },
  ],
  "exit-up": [
    { opacity: "1", transform: "translateY(0)" },
    { opacity: "1", transform: "translateY(-100%)" },
  ],
};

/** Static first-frame pins for kinds whose animation starts on a later frame. */
const KIND_PINS: Record<Exclude<LayerKind, "base">, { opacity: string; transform: string }> = {
  enter: { opacity: "0", transform: "translateY(24px)" },
  "cover-down": { opacity: "1", transform: "translateY(-100%)" },
  "exit-up": { opacity: "1", transform: "translateY(0)" },
};

/**
 * Play a layer's animation for its current kind. Called from a reactive
 * effect (below), so a reused DOM node re-animates when the swap flips its
 * kind — that is how the exit-up slide runs on a node that mounted long
 * before. As before, the animation must start only once the element is
 * attached (a pending animation never starts otherwise), and the first
 * frame is pinned inline until it takes over.
 */
function playKind(ref: HTMLDivElement, kind: LayerKind, swap: number | undefined): void {
  if (kind === "base") return;
  const frames = KIND_FRAMES[kind];
  if (typeof ref.animate !== "function") return; // non-browser environment (tests)

  // Kill any CSS animation before it can start; this does not touch WAAPI.
  ref.style.animation = "none";
  // Pin the first keyframe until the animation takes over: the frame spent
  // in the DOM before the deferred `animate` runs must not show the end
  // state (that reads as a flash of the new article).
  ref.style.opacity = KIND_PINS[kind].opacity;
  ref.style.transform = KIND_PINS[kind].transform;
  requestAnimationFrame(() => {
    if (!ref.isConnected) return;
    ref.style.opacity = "";
    ref.style.transform = "";
    const animation = ref.animate(frames, { duration: COVER_MS, easing: "ease-out" });
    // The slide is over: let the stack drop what this swap left behind.
    // A rejection means the layer was replaced mid-flight; the newer swap
    // rebuilds the stack anyway.
    if (swap !== undefined)
      animation.finished.then(
        () => dropFinished(swap),
        () => {},
      );
  });
}

function ArticleLayer(props: { layer: Layer }) {
  let ref: HTMLDivElement | undefined;
  // The node is reused across swaps (the same layer object flips kind), so
  // the animation tracks the kind reactively instead of running only on
  // mount — that is what makes the exit-up slide play on a mounted node.
  createEffect(() => {
    const kind = props.layer.kind;
    if (ref) playKind(ref, kind, props.layer.swap);
  });
  const bindLayer = (el: HTMLDivElement) => {
    ref = el;
    playKind(el, props.layer.kind, props.layer.swap);
  };

  return (
    <div
      ref={bindLayer}
      data-layer={props.layer.kind}
      class={cn(
        "flex min-h-0 flex-1 flex-col",
        props.layer.kind === "base" && "pointer-events-none absolute inset-0 z-0",
        props.layer.kind === "enter" &&
          "animate-in duration-300 ease-out fade-in slide-in-from-bottom-6",
        props.layer.kind === "cover-down" &&
          "absolute inset-0 z-10 animate-in bg-background duration-300 ease-out slide-in-from-top-full",
        props.layer.kind === "exit-up" && "absolute inset-0 z-10 bg-background",
      )}
    >
      <header class="border-b px-6 py-4">
        <div class="flex items-start gap-3">
          <div class="min-w-0 flex-1">
            <h1 class="text-lg leading-snug font-semibold">{props.layer.post.title}</h1>
            <div class="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-[0.6875rem] text-muted-foreground">
              <span class="tabular-nums">{formatDateTime(props.layer.post.publishedAt)}</span>
              <Show when={props.layer.post.author}>
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
              label={props.layer.post.starred ? "取消星标" : "加星标"}
              onClick={() => void toggleStar(props.layer.post.guid)}
            >
              <Star class={cn("size-4", props.layer.post.starred && "fill-current text-warning")} />
            </IconButton>
            <IconButton
              label={props.layer.post.read ? "标记为未读" : "标记为已读"}
              onClick={() => void setOpenPostRead(!props.layer.post.read)}
            >
              <MailOpen class="size-4" />
            </IconButton>
            <IconButton
              label="在浏览器中打开原文"
              disabled={!props.layer.post.url}
              onClick={() => void openLink(props.layer.post.url)}
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
          innerHTML={sanitizeHTML(
            props.layer.post.content || props.layer.post.summary,
            props.layer.post.url,
          )}
        />
      </div>
    </div>
  );
}
