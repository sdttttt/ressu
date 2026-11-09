import { toast } from "solid-sonner";

import { applyFontSize, applyTheme } from "~/lib/theme";
import type { FeedsSnapshot, SyncState } from "~/mygo";
import { app as state, idleSync, setAppState as setState } from "~/stores/app-state";
import { syncFeedIcons } from "~/stores/icons";
import { loadPosts } from "~/stores/posts";

/** Applies the feed list the Go side sent. */
export async function replaceFeeds(snapshot: FeedsSnapshot): Promise<void> {
  syncFeedIcons(state.feeds, snapshot.feeds);
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

/** Refreshes the list of the selected feed after its posts changed. */
export function handlePostsChanged(feedId: number): void {
  if (feedId === state.current.feedId || state.current.feedId === null) void loadPosts(0);
}
