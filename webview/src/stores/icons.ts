import { Feeds } from "~/mygo";
import type { Feed } from "~/mygo";
import { app as state, setAppState as setState } from "~/stores/app-state";
import type { FeedIconView } from "~/stores/app-state";

/** How long a locally rendered icon can be reused before Go revalidates it. */
const ICON_TTL_MS = 7 * 24 * 60 * 60 * 1000;

/** In-flight icon loads, keyed by feed id. */
const requests = new Map<number, Promise<void>>();

/** Starts a local icon load for a feed unless it is already present or known
 * to be unavailable. Concurrent calls share one request. */
export async function ensureFeedIcon(feedID: number): Promise<void> {
  const current = requests.get(feedID);
  if (current) {
    await current;
    return;
  }

  const cached = state.icons[feedID];
  const feed = state.feeds.find((item) => item.id === feedID);
  if (cached?.missing || !feed) return;
  if (cached && !iconStale(cached)) return;

  const request = loadFeedIcon(feedID, cached);
  requests.set(feedID, request);
  try {
    await request;
  } finally {
    requests.delete(feedID);
  }
}

/** Updates cached icons when a feed changes and removes entries for feeds that
 * are no longer subscribed. */
export function syncFeedIcons(previous: Feed[], next: Feed[]): void {
  const kept = new Set(next.map((feed) => feed.id));
  setState("icons", (icons) => {
    let changed = false;
    const updated: Record<number, FeedIconView> = {};
    for (const [id, icon] of Object.entries(icons)) {
      const feedID = Number(id);
      if (kept.has(feedID)) updated[feedID] = icon;
      else changed = true;
    }

    for (const feed of next) {
      const old = previous.find((item) => item.id === feed.id);
      if (old && old.iconUrl !== feed.iconUrl) delete updated[feed.id];
    }
    return changed || Object.keys(updated).length !== Object.keys(icons).length ? updated : icons;
  });
}

/** Removes a feed's locally cached icon. */
export function forgetFeedIcon(feedID: number): void {
  requests.delete(feedID);
  setState("icons", (icons) => {
    if (!(feedID in icons)) return icons;
    const updated = { ...icons };
    delete updated[feedID];
    return updated;
  });
}

/** Loads one icon through Go and stores a data URL in local state. */
async function loadFeedIcon(feedID: number, cached?: FeedIconView): Promise<void> {
  if (!cached) {
    setIcon(feedID, { src: "", sourceURL: "", fetchedAt: "", missing: true });
  }

  try {
    const icon = await Feeds.icon(feedID);
    if (!icon.data) throw new Error("empty icon");
    setIcon(feedID, {
      src: `data:${icon.meta.contentType};base64,${icon.data}`,
      sourceURL: icon.meta.sourceUrl,
      fetchedAt: icon.meta.fetchedAt,
      missing: false,
    });
  } catch {
    setIcon(feedID, { src: "", sourceURL: "", fetchedAt: "", missing: true });
  }
}

/** Writes one icon view into the Solid store. */
function setIcon(feedID: number, icon: FeedIconView): void {
  setState("icons", feedID, icon);
}

/** Reports whether a cached icon is absent from the local state or old enough
 * to ask Go for a possibly refreshed copy. */
function iconStale(icon: FeedIconView): boolean {
  if (!icon.src || icon.missing) return true;
  const fetchedAt = Date.parse(icon.fetchedAt);
  if (Number.isNaN(fetchedAt)) return true;
  return Date.now() - fetchedAt >= ICON_TTL_MS;
}
