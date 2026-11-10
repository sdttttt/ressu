import { toast } from "~/lib/notifications";
import { Feeds, Status } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { syncFeedIcons } from "~/stores/icons";
import { loadPosts } from "~/stores/posts";
import { closeSnapshot } from "~/stores/snapshot-actions";

/** Selects a subscription, or deselects it to show the aggregate list. */
export async function selectFeed(id: number): Promise<void> {
  closeSnapshot();
  const selected = state.current.feedId === id;
  void Status.selectFeed(selected ? 0 : id).catch(() => undefined);
  setState("current", { feedId: selected ? null : id, postFeedId: null, guid: null });
  setState("post", null);
  await loadPosts(0);
}

/** Fetches one feed. */
export async function refreshFeed(feedId: number): Promise<void> {
  try {
    const previous = state.feeds.find((feed) => feed.id === feedId);
    const feed = await Feeds.refresh(feedId);
    syncFeedIcons(previous ? [previous] : [], [feed]);
    setState("feeds", (item) => item.id === feed.id, feed);
  } catch (error) {
    toast.error(`${feedId}：${messageOf(error)}`);
  }
}

/** Refreshes every subscription. */
export async function refreshAll(): Promise<void> {
  if (state.sync.active) return;

  try {
    const summary = await Feeds.refreshAll();
    if (summary.failed > 0) {
      toast.warning(`${summary.updated} 个订阅源已更新，${summary.failed} 个失败`);
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Opens the native add-feed window. */
export async function openAddFeed(): Promise<void> {
  try {
    await Feeds.openAddFeed();
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Imports subscriptions from an OPML file through a native file picker. */
export async function importOPML(): Promise<void> {
  try {
    const summary = await Feeds.importOPML();
    if (summary.failed > 0) {
      toast.warning(
        `${summary.imported} 个已导入，${summary.skipped} 个跳过，${summary.failed} 个失败`,
      );
    } else if (summary.skipped > 0) {
      toast.info(`${summary.imported} 个已导入，${summary.skipped} 个已存在`);
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Exports every subscription to an OPML file through a native save dialog. */
export async function exportOPML(): Promise<void> {
  try {
    const path = await Feeds.exportOPML();
    if (path) {
      toast.success(`已导出 ${state.feeds.length} 个订阅源`);
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Confirms and unsubscribes from a feed. */
export async function removeFeed(id: number): Promise<void> {
  try {
    await Feeds.confirmRemove(id);
  } catch (error) {
    toast.error(messageOf(error));
  }
}
