import { ExternalLink, MailOpen, Newspaper, Star } from "lucide-solid";
import { Show, batch, createEffect, createSignal, onCleanup, onMount, untrack } from "solid-js";
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

/** How the current article entered the screen. */
type Entry = "next" | "prev" | "enter";

function handleLinkClick(event: MouseEvent, root: HTMLElement): void {
  const target = event.target as HTMLElement | null;
  const anchor = target?.closest("a");
  if (!anchor || !root.contains(anchor)) return;

  const href = anchor.getAttribute("href");
  if (!href) return;
  event.preventDefault();
  void openLink(href);
}

/** How the new article should cover the old one, based on list order. */
function coverDirection(from: Post | null, to: Post): "next" | "prev" {
  const indexAt = (guid: string) => app.posts.findIndex((post) => post.guid === guid);
  const fromIndex = from ? indexAt(from.guid) : -1;
  const toIndex = indexAt(to.guid);
  if (fromIndex === -1 || toIndex === -1) return "next";
  return toIndex > fromIndex ? "next" : "prev";
}

/**
 * Holds the articles on screen one step behind the store: when another post is
 * opened, the old one stays perfectly still and the new one slides over it
 * like a sheet of paper — from above when reading further down the list, from
 * below when going back up. The first article just fades in, since there is
 * nothing to cover. The covered article is dropped once the slide is done.
 * Star and read toggles on the open post update in place and never restart
 * the cover.
 */
function createArticleFade(): {
  current: () => string;
  leaving: () => string;
  entry: () => Entry;
  post: () => Post | null;
  leavingPost: () => Post | null;
} {
  const [current, setCurrent] = createSignal("");
  const [leaving, setLeaving] = createSignal("");
  const [entry, setEntry] = createSignal<Entry>("enter");
  // Field-level reactive snapshots: plain signals would either hold a store
  // proxy (which `setState` merges in place, silently mutating between runs)
  // or restart this effect with every fresh copy written into it.
  const [snap, setSnap] = createStore<{ post: Post | null; leavingPost: Post | null }>({
    post: null,
    leavingPost: null,
  });
  let dropTimer: ReturnType<typeof setTimeout> | undefined;

  createEffect(() => {
    const latest = app.post;

    if (!latest) {
      if (current()) {
        batch(() => {
          setCurrent("");
          setLeaving("");
          setEntry("enter");
          setSnap({ post: null, leavingPost: null });
        });
      }
      return;
    }
    if (!current() || latest.guid === current()) {
      setSnap("post", reconcile({ ...latest }));
      setCurrent(latest.guid);
      return;
    }

    // untrack: the effect writes what it reads here, so a tracked read would
    // rerun it forever.
    const outgoing = untrack(() => snap.post);
    const hasCover = Boolean(outgoing);
    batch(() => {
      setSnap({
        post: { ...latest },
        leavingPost: outgoing ? { ...outgoing } : null,
      });
      setLeaving(outgoing?.guid ?? "");
      setEntry(hasCover ? coverDirection(outgoing, latest) : "enter");
    });
    clearTimeout(dropTimer);
    dropTimer = setTimeout(() => setLeaving(""), COVER_MS);
  });

  onCleanup(() => clearTimeout(dropTimer));

  return {
    current,
    leaving,
    entry,
    post: () => snap.post,
    leavingPost: () => snap.leavingPost,
  };
}

export default function PostView() {
  let articleRef: HTMLElement | undefined;

  const fade = createArticleFade();

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
        when={fade.current()}
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
              moved in the list. */}
          <Show keyed when={fade.leaving()}>
            {(guid) => <ArticleLayer key={guid} layer="base" post={fade.leavingPost} />}
          </Show>
          <Show keyed when={fade.current()}>
            {(guid) => <ArticleLayer key={guid} layer={fade.entry()} post={fade.post} />}
          </Show>
        </div>
      </Show>
    </article>
  );
}

function ArticleLayer(props: { key: string; layer: Entry | "base"; post: () => Post | null }) {
  // The key is unused in markup, but ties the layer instance to its guid so a
  // swap replaces the element instead of updating it in place.
  void props.key;

  return (
    <div
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
            <h1 class="text-lg leading-snug font-semibold">{props.post()?.title}</h1>
            <div class="mt-1.5 flex flex-wrap items-center gap-x-2 gap-y-1 text-[0.6875rem] text-muted-foreground">
              <span class="tabular-nums">{formatDateTime(props.post()?.publishedAt ?? "")}</span>
              <Show when={props.post()?.author}>
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
              label={props.post()?.starred ? "取消星标" : "加星标"}
              onClick={() => {
                const guid = props.post()?.guid;
                if (guid) void toggleStar(guid);
              }}
            >
              <Star class={cn("size-4", props.post()?.starred && "fill-current text-warning")} />
            </IconButton>
            <IconButton
              label={props.post()?.read ? "标记为未读" : "标记为已读"}
              onClick={() => {
                const post = props.post();
                if (post) void setOpenPostRead(!post.read);
              }}
            >
              <MailOpen class="size-4" />
            </IconButton>
            <IconButton
              label="在浏览器中打开原文"
              disabled={!props.post()?.url}
              onClick={() => {
                const url = props.post()?.url;
                if (url) void openLink(url);
              }}
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
            props.post()?.content || props.post()?.summary || "",
            props.post()?.url ?? "",
          )}
        />
      </div>
    </div>
  );
}
