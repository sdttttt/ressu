import { toast } from "~/lib/notifications";
import { Feeds, Status } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
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
