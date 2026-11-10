import { toast } from "solid-sonner";

import { externalHref } from "~/lib/sanitize";
import { applyFontSize, applyTheme } from "~/lib/theme";
import { Settings, type Prefs } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import { messageOf } from "~/stores/errors";

/** Saves the preferences, applying what they change. */
export async function updatePrefs(patch: Partial<Prefs>): Promise<void> {
  try {
    const prefs = await Settings.update({ ...state.prefs, ...patch });
    setState("prefs", prefs);
    if (patch.theme !== undefined) applyTheme(prefs.theme);
    if (patch.fontSize !== undefined) applyFontSize(prefs.fontSize);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

let paneSizesTimer: ReturnType<typeof setTimeout> | undefined;

type PaneSizeOptions = { sidebarCollapsed?: boolean };

/** Saves the widths of the panes once the user stops dragging. */
export function savePaneSizes(sizes: number[], options: PaneSizeOptions = {}): void {
  const [first, second] = sizes;
  if (first === undefined || second === undefined) return;

  const sidebarPercent = options.sidebarCollapsed
    ? state.prefs.paneSizes[0]
    : Math.round(first * 100);
  const listRatio = options.sidebarCollapsed ? Math.min(Math.max(first, 0.18), 0.55) : second;
  const listPercent = options.sidebarCollapsed
    ? Math.round(listRatio * (100 - sidebarPercent))
    : Math.round(listRatio * 100);
  const paneSizes = [sidebarPercent, listPercent];
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

/** Opens the native settings window. */
export async function openSettings(): Promise<void> {
  try {
    await Settings.openSettings();
  } catch (error) {
    toast.error(messageOf(error));
  }
}
