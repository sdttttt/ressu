import { toast } from "solid-sonner";

import { applyFontSize, applyTheme } from "~/lib/theme";
import { Feeds, Settings } from "~/mygo";
import { setAppState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { loadPosts } from "~/stores/posts";

/** Loads the subscriptions and the preferences, then shows all posts. */
export async function start(): Promise<void> {
  try {
    const [feeds, prefs, dataDir] = await Promise.all([
      Feeds.list(),
      Settings.get(),
      Settings.dataDir(),
    ]);
    setAppState({
      feeds,
      prefs,
      dataDir,
      ready: true,
      sidebar: { collapsed: prefs.sidebarCollapsed, visible: false },
    });
    applyTheme(prefs.theme);
    applyFontSize(prefs.fontSize);

    await loadPosts(0);
  } catch (error) {
    setAppState("ready", true);
    toast.error(`无法读取本地数据：${messageOf(error)}`);
  }
}
