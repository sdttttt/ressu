import { toast } from "solid-sonner";

import { generateSnapshot } from "~/lib/snapshot";
import { Posts } from "~/mygo";
import { app as state, currentFeed, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";

/** Captures the open post, stitches it and opens the preview dialog. */
export async function captureSnapshot(): Promise<void> {
  const post = state.post;
  if (!post || state.snapshotting) return;
  const feedTitle = currentFeed()?.title ?? "";

  setState({ snapshotting: true, snapshot: null, snapshotAction: null });
  try {
    const tiles = await generateSnapshot(post, feedTitle);
    const png = await Posts.stitchSnapshot(tiles);
    setState({ snapshotting: false, snapshot: { post, png } });
  } catch (error) {
    setState({ snapshotting: false, snapshot: null });
    toast.error(`无法生成快照：${messageOf(error)}`);
  } finally {
    setState("snapshotting", false);
  }
}

/** Closes the snapshot preview and releases its action state. */
export function closeSnapshot(): void {
  setState({ snapshot: null, snapshotAction: null });
}

/** Saves the previewed snapshot as a PNG through a save dialog. */
export async function saveSnapshot(): Promise<void> {
  const snapshot = state.snapshot;
  if (!snapshot || state.snapshotAction) return;

  setState("snapshotAction", "save");
  try {
    const path = await Posts.saveSnapshot(snapshot.post.title, snapshot.png);
    if (path) {
      toast.success(`快照已保存：${path}`);
      closeSnapshot();
    }
  } catch (error) {
    toast.error(`无法保存快照：${messageOf(error)}`);
  } finally {
    setState("snapshotAction", null);
  }
}

/** Puts the previewed snapshot on the clipboard as a PNG. */
export async function copySnapshot(): Promise<void> {
  const snapshot = state.snapshot;
  if (!snapshot || state.snapshotAction) return;

  setState("snapshotAction", "copy");
  try {
    await Posts.copySnapshot(snapshot.png);
    toast.success("快照已复制到剪贴板");
    closeSnapshot();
  } catch (error) {
    toast.error(`无法复制快照：${messageOf(error)}`);
  } finally {
    setState("snapshotAction", null);
  }
}
