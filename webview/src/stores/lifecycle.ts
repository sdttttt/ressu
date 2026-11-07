import { toast } from "solid-sonner";

import { applyFontSize, applyTheme } from "~/lib/theme";
import { Feeds, Settings } from "~/mygo";
import { setAppState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { selectFeed } from "~/stores/feeds";

/** Loads the subscriptions and the preferences, and selects the first feed. */
export async function start(): Promise<void> {
  try {
    const [feeds, prefs, dataDir] = await Promise.all([
      Feeds.list(),
      Settings.get(),
      Settings.dataDir(),
    ]);
    setAppState({ feeds, prefs, dataDir, ready: true });
    applyTheme(prefs.theme);
    applyFontSize(prefs.fontSize);

    const first = feeds[0];
    if (first) await selectFeed(first.id);
  } catch (error) {
    setAppState("ready", true);
    toast.error(`无法读取本地数据：${messageOf(error)}`);
  }
}
