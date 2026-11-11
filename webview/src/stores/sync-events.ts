import { toast } from "~/lib/notifications";
import { applyFontSize, applyTheme } from "~/lib/theme";
import type { FeedsSnapshot, PostSelection, PostSummary, Prefs, SyncState } from "~/mygo";
import { app as state, idleSync, setAppState as setState } from "~/stores/app-state";
import { loadPosts, selectPost } from "~/stores/posts";

/** Applies the feed list the Go side sent. */
export async function replaceFeeds(snapshot: FeedsSnapshot): Promise<void> {
  setState({ feeds: snapshot.feeds, totalUnread: snapshot.unread });
  if (
    state.current.feedId !== null &&
    !snapshot.feeds.some((feed) => feed.id === state.current.feedId)
  ) {
    setState({ current: { feedId: null, postFeedId: null, guid: null }, post: null });
  }

  if (state.current.feedId === null) await loadPosts(0);
}

/** Applies a refresh event. */
export function applySync(event: SyncState): void {
  if (event.active) {
    setState("sync", event);
    if (event.error) {
      setState("syncErrors", (errors) => [...errors, `${event.feed || "订阅源"}：${event.error}`]);
    }
    return;
  }

  setState("sync", { ...idleSync, done: event.total, total: event.total, error: undefined });
  if (state.syncErrors.length > 0) {
    toast.error(`${state.syncErrors.length} 个订阅源刷新失败，详情见状态栏`);
    setState("syncErrors", []);
  }
}

/** Starts a new refresh run and clears errors from the previous one. */
export function beginSync(event: SyncState): void {
  setState("syncErrors", []);
  setState("sync", event);
}

/** Applies the sync section of a full state snapshot without ending a run. */
export function applySyncSnapshot(event: SyncState): void {
  setState("sync", event);
}

/** Applies theme and font size from another window or native command. */
export function applyAppearance(theme: string, fontSize: string): void {
  setState("prefs", "theme", theme);
  setState("prefs", "fontSize", fontSize);
  applyTheme(theme);
  applyFontSize(fontSize);
}

/** Applies preferences changed by the native settings window or another page. */
export async function applyPrefs(prefs: Prefs): Promise<void> {
  const postsPerFeedChanged = prefs.postsPerFeed !== state.prefs.postsPerFeed;
  setState("prefs", prefs);
  applyTheme(prefs.theme);
  applyFontSize(prefs.fontSize);
  if (postsPerFeedChanged) await loadPosts(0);
}

/** Refreshes the list of the selected feed after its posts changed. */
export function handlePostsChanged(feedId: number): void {
  if (feedId === state.current.feedId || state.current.feedId === null) void loadPosts(0);
}

/** Applies another page's selected subscription. */
export async function handleSelectionChanged(feedId: number): Promise<void> {
  if (state.current.feedId === feedId) return;

  setState("current", { feedId, postFeedId: null, guid: null });
  setState("post", null);
  await loadPosts(0);
}

/** Opens the article selected by the native list. */
export async function handlePostSelected(selection: PostSelection): Promise<void> {
  if (state.current.postFeedId === selection.feedId && state.current.guid === selection.guid) {
    return;
  }
  await selectPost(selection.guid, selection.feedId);
}

/** Applies article read and starred changes to the list and open reader. */
export function handlePostUpdated(post: PostSummary): void {
  const index = state.posts.findIndex(
    (item) => item.feedId === post.feedId && item.guid === post.guid,
  );
  if (index >= 0) {
    const previous = state.posts[index];
    if (previous.read !== post.read) {
      setState("postUnread", (unread) => Math.max(unread + (post.read ? -1 : 1), 0));
    }
    setState("posts", index, "read", post.read);
    setState("posts", index, "starred", post.starred);
  }

  if (state.post?.feedId === post.feedId && state.post.guid === post.guid) {
    setState("post", (current) => (current ? { ...current, ...post } : current));
  }
}
