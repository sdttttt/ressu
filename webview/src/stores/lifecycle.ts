import { toast } from "~/lib/notifications";
import { applyFontSize, applyTheme } from "~/lib/theme";
import { Feeds, Settings, Status } from "~/mygo";
import { setAppState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { loadPosts } from "~/stores/posts";

/** Loads the subscriptions and the preferences, then shows all posts. */
export async function start(): Promise<void> {
  void Status.pageReady().catch(() => undefined);
  try {
    const [snapshot, prefs] = await Promise.all([Feeds.list(), Settings.get()]);
    setAppState({
      feeds: snapshot.feeds,
      totalUnread: snapshot.unread,
      prefs,
      ready: true,
    });
    applyTheme(prefs.theme);
    applyFontSize(prefs.fontSize);

    await loadPosts(0);
  } catch (error) {
    setAppState("ready", true);
    toast.error(`无法读取本地数据：${messageOf(error)}`);
  }
}
