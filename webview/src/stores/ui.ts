import { app, setAppState, type DialogName } from "~/stores/app-state";
import { updatePrefs } from "~/stores/settings";

export function openDialog(name: DialogName): void {
  setAppState("dialog", name, true);
}

export function closeDialog(name: DialogName): void {
  setAppState("dialog", name, false);
}

export function setFeedFilter(value: string): void {
  setAppState("feedFilter", value);
}

const SIDEBAR_SHOW_DELAY_MS = 180;
const SIDEBAR_HIDE_DELAY_MS = 200;
const SIDEBAR_RETRIGGER_INTERVAL_MS = 500;
const SIDEBAR_INITIAL_SHOW_DELAY_MS = 30;

let sidebarShowTimer: ReturnType<typeof setTimeout> | undefined;
let sidebarHideTimer: ReturnType<typeof setTimeout> | undefined;
let sidebarShownAt = 0;

function clearSidebarTimers(): void {
  clearTimeout(sidebarShowTimer);
  clearTimeout(sidebarHideTimer);
  sidebarShowTimer = undefined;
  sidebarHideTimer = undefined;
}

function scheduleSidebarShow(delay: number): void {
  if (sidebarShowTimer !== undefined) return;

  sidebarShowTimer = setTimeout(() => {
    sidebarShowTimer = undefined;
    showSidebar();
  }, delay);
}

/** Turns the docked feed sidebar into its floating overlay. */
export async function collapseSidebar(): Promise<void> {
  clearSidebarTimers();
  setAppState("sidebar", { collapsed: true, visible: false });
  scheduleSidebarShow(SIDEBAR_INITIAL_SHOW_DELAY_MS);
  await updatePrefs({ sidebarCollapsed: true });
}

/** Restores the docked three-pane layout. */
export async function expandSidebar(): Promise<void> {
  clearSidebarTimers();
  setAppState("sidebar", { collapsed: false, visible: false });
  await updatePrefs({ sidebarCollapsed: false });
}

/** Shows the floating feed sidebar while the docked sidebar stays hidden. */
function showSidebar(): void {
  clearSidebarTimers();
  setAppState("sidebar", "visible", true);
  sidebarShownAt = Date.now();
}

/** Reveals the sidebar only after the cursor shows hover intent. */
export function requestSidebar(): void {
  clearTimeout(sidebarHideTimer);
  sidebarHideTimer = undefined;
  if (app.sidebar.visible || sidebarShowTimer !== undefined) return;

  const cooldown = Math.max(0, SIDEBAR_RETRIGGER_INTERVAL_MS - (Date.now() - sidebarShownAt));
  scheduleSidebarShow(SIDEBAR_SHOW_DELAY_MS + cooldown);
}

/** Keeps the sidebar shown when the cursor returns before the hide delay ends. */
export function keepSidebar(): void {
  clearTimeout(sidebarHideTimer);
  sidebarHideTimer = undefined;
}

/** Slides the floating feed sidebar back behind the left edge. */
export function hideSidebar(): void {
  clearTimeout(sidebarShowTimer);
  sidebarShowTimer = undefined;
  if (!app.sidebar.visible || sidebarHideTimer !== undefined) return;

  sidebarHideTimer = setTimeout(() => {
    sidebarHideTimer = undefined;
    setAppState("sidebar", "visible", false);
  }, SIDEBAR_HIDE_DELAY_MS);
}
