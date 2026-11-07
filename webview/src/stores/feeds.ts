import { toast } from "solid-sonner";

import { Feeds } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { loadPosts } from "~/stores/posts";
import { closeSnapshot } from "~/stores/snapshot-actions";
import { closeDialog } from "~/stores/ui";

/** Selects a subscription and loads the first page of its posts. */
export async function selectFeed(id: number): Promise<void> {
  closeSnapshot();
  setState("current", { feedId: id, guid: null });
  setState("post", null);
  await loadPosts(0);
}

/** Fetches one feed. */
export async function refreshFeed(feedId: number): Promise<void> {
  try {
    const feed = await Feeds.refresh(feedId);
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
    if (summary.updated + summary.failed === 0) {
      toast.info("还没有订阅源");
    } else if (summary.failed > 0) {
      toast.warning(`${summary.updated} 个订阅源已更新，${summary.failed} 个失败`);
    } else if (summary.new > 0) {
      toast.success(`新增 ${summary.new} 篇文章`);
    } else {
      toast.success("所有订阅源都已是最新");
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Subscribes to a feed. Returns whether it worked. */
export async function addFeed(feedURL: string): Promise<boolean> {
  try {
    const feed = await Feeds.add(feedURL.trim());
    setState("feeds", (feeds) => [...feeds.filter((item) => item.id !== feed.id), feed]);
    closeDialog("addFeed");
    toast.success(`已添加「${feed.title}」`);
    await selectFeed(feed.id);
    return true;
  } catch (error) {
    toast.error(messageOf(error));
    return false;
  }
}

/** Imports subscriptions from an OPML file through a native file picker. */
export async function importOPML(): Promise<void> {
  try {
    const summary = await Feeds.importOPML();
    if (summary.imported + summary.skipped + summary.failed === 0) {
      toast.info("没有选择文件");
      return;
    }
    if (summary.failed > 0) {
      toast.warning(
        `${summary.imported} 个已导入，${summary.skipped} 个跳过，${summary.failed} 个失败`,
      );
    } else if (summary.skipped > 0) {
      toast.info(`${summary.imported} 个已导入，${summary.skipped} 个已存在`);
    } else {
      toast.success(`已导入 ${summary.imported} 个订阅源`);
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Unsubscribes from a feed, selecting another one. */
export async function removeFeed(id: number): Promise<void> {
  const wasCurrent = state.current.feedId === id;
  try {
    await Feeds.remove(id);
    const feeds = state.feeds.filter((feed) => feed.id !== id);
    setState("feeds", feeds);
    if (!wasCurrent) return;

    const next = feeds[0];
    if (next) {
      await selectFeed(next.id);
    } else {
      setState({
        posts: [],
        postTotal: 0,
        postUnread: 0,
        post: null,
        current: { feedId: null, guid: null },
      });
    }
    toast.success("已取消订阅");
  } catch (error) {
    toast.error(messageOf(error));
  }
}
