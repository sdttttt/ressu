import { toast } from "solid-sonner";

import type { Feed, SyncState } from "~/mygo";
import { app as state, idleSync, setAppState as setState } from "~/stores/app-state";
import { syncFeedIcons } from "~/stores/icons";
import { loadPosts } from "~/stores/posts";

/** Applies the feed list the Go side sent. */
export async function replaceFeeds(feeds: Feed[]): Promise<void> {
  syncFeedIcons(state.feeds, feeds);
  setState("feeds", feeds);
  if (state.current.feedId !== null && !feeds.some((feed) => feed.id === state.current.feedId)) {
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

/** Refreshes the list of the selected feed after its posts changed. */
export function handlePostsChanged(feedId: number): void {
  if (feedId === state.current.feedId || state.current.feedId === null) void loadPosts(0);
}
