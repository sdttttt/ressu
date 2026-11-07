import { toast } from "solid-sonner";

import { externalHref } from "~/lib/sanitize";
import { applyFontSize, applyTheme } from "~/lib/theme";
import { Settings, type Prefs } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";
import { loadPosts } from "~/stores/posts";

/** Saves the preferences, applying what they change. */
export async function updatePrefs(patch: Partial<Prefs>): Promise<void> {
  try {
    const prefs = await Settings.update({ ...state.prefs, ...patch });
    setState("prefs", prefs);
    if (patch.theme !== undefined) applyTheme(prefs.theme);
    if (patch.fontSize !== undefined) applyFontSize(prefs.fontSize);
    if (patch.postsPerFeed !== undefined && state.current.feedId !== null) {
      await loadPosts(0);
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

let paneSizesTimer: ReturnType<typeof setTimeout> | undefined;

/** Saves the widths of the panes once the user stops dragging. */
export function savePaneSizes(sizes: number[]): void {
  const [sidebar, list] = sizes;
  if (sidebar === undefined || list === undefined) return;

  const paneSizes = [Math.round(sidebar * 100), Math.round(list * 100)];
  const [wasSidebar, wasList] = state.prefs.paneSizes;
  if (paneSizes[0] === wasSidebar && paneSizes[1] === wasList) return;

  clearTimeout(paneSizesTimer);
  paneSizesTimer = setTimeout(() => {
    void updatePrefs({ paneSizes });
  }, 500);
}

/** Opens a link outside the reader, the way the preferences ask for. */
export async function openLink(href: string): Promise<void> {
  const url = externalHref(href);
  if (!url) return;
  try {
    await Settings.openLink(url);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Shows the data directory in the file manager. */
export async function openDataDir(): Promise<void> {
  try {
    await Settings.openDataDir();
  } catch (error) {
    toast.error(messageOf(error));
  }
}
