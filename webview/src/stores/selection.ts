import { app as state } from "~/stores/app-state";
import { selectFeed } from "~/stores/feeds";
import { selectPost, setOpenPostRead } from "~/stores/posts";
import { openLink } from "~/stores/settings";

/** The list index the selection would move to, honouring the unread-only
 * filter. Returns -1 when there is nothing to move onto. */
function selectionTarget(delta: 1 | -1, onlyUnread: boolean): number {
  const posts = state.posts;
  if (posts.length === 0) return -1;

  const current = posts.findIndex(
    (post) => post.feedId === state.current.postFeedId && post.guid === state.current.guid,
  );
  let index = current === -1 ? (delta === 1 ? 0 : posts.length - 1) : current + delta;

  while (index >= 0 && index < posts.length) {
    if (!onlyUnread || !posts[index]!.read) return index;
    index += delta;
  }
  return -1;
}

/** Scrolls the post list so its row is visible. */
function revealPost(feedId: number, guid: string): void {
  document
    .querySelector<HTMLElement>(
      `[data-feed-id="${CSS.escape(String(feedId))}"][data-guid="${CSS.escape(guid)}"]`,
    )
    ?.scrollIntoView({ block: "nearest" });
}

/**
 * Moves the selection within the loaded posts, optionally skipping read ones,
 * and opens the post it lands on. Stops at the ends of the list instead of
 * wrapping around.
 */
export async function moveSelection(delta: 1 | -1, onlyUnread = false): Promise<void> {
  const index = selectionTarget(delta, onlyUnread);
  if (index === -1) return;

  const post = state.posts[index]!;
  await selectPost(post.guid, post.feedId);
  revealPost(post.feedId, post.guid);
}

/** Opens the selected post, or the first post when nothing is selected. */
export async function openSelectedPost(): Promise<void> {
  const index = selectionTarget(1, false);
  if (index === -1) return;

  const post = state.posts[index]!;
  await selectPost(post.guid, post.feedId);
  revealPost(post.feedId, post.guid);
}

/** Opens the selected post's original page in a browser. */
export async function openSelectedPostInBrowser(): Promise<void> {
  const post = state.post;
  if (post) await openLink(post.url);
}

/** Toggles the read state of the open post. */
export async function toggleOpenPostRead(): Promise<void> {
  const post = state.post;
  if (post) await setOpenPostRead(!post.read);
}

/** Selects the feed next to or before the selected one. */
export async function selectNeighbourFeed(delta: 1 | -1): Promise<void> {
  const feeds = state.feeds;
  if (feeds.length === 0) return;

  const current = feeds.findIndex((feed) => feed.id === state.current.feedId);
  const index = current === -1 ? (delta === 1 ? 0 : feeds.length - 1) : current + delta;
  const feed = feeds[index];
  if (feed) await selectFeed(feed.id);
}
