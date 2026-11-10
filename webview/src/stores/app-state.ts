import { createStore } from "solid-js/store";

import type { Feed, Post, PostSummary, Prefs, SortOrder, SyncState } from "~/mygo";

export type DialogName = "help";

type SnapshotAction = "copy" | "save";

type Snapshot = {
  post: Post;
  png: string;
};

export type FeedIconView = {
  src: string;
  sourceURL: string;
  fetchedAt: string;
  missing: boolean;
};

type AppState = {
  feeds: Feed[];
  totalUnread: number;
  icons: Record<number, FeedIconView>;
  posts: PostSummary[];
  postTotal: number;
  postUnread: number;
  post: Post | null;
  prefs: Prefs;
  sync: SyncState;
  syncErrors: string[];
  ready: boolean;
  loadingPosts: boolean;
  loadingPost: boolean;
  snapshotting: boolean;
  snapshot: Snapshot | null;
  snapshotAction: SnapshotAction | null;
  unreadOnly: boolean;
  starredOnly: boolean;
  postSearch: string;
  postSearchOpen: boolean;
  postSortOrder: SortOrder;
  current: { feedId: number | null; postFeedId: number | null; guid: string | null };
  dialog: Record<DialogName, boolean>;
};

const defaultPrefs: Prefs = {
  theme: "system",
  fontSize: "medium",
  refreshOnStart: true,
  markReadOnOpen: true,
  remoteResources: "goProxy",
  openLinks: "browser",
  postsPerFeed: 500,
  imageCacheLimit: 256,
  paneSizes: [18, 30],
  sidebarCollapsed: false,
};

/** The sync state before a refresh starts or after one finishes. */
export const idleSync: SyncState = { active: false, done: 0, total: 0, feedId: 0, feed: "" };

const [state, setState] = createStore<AppState>({
  feeds: [],
  totalUnread: 0,
  icons: {},
  posts: [],
  postTotal: 0,
  postUnread: 0,
  post: null,
  prefs: defaultPrefs,
  sync: { ...idleSync },
  syncErrors: [],
  ready: false,
  loadingPosts: false,
  loadingPost: false,
  snapshotting: false,
  snapshot: null,
  snapshotAction: null,
  unreadOnly: false,
  starredOnly: false,
  postSearch: "",
  postSearchOpen: false,
  postSortOrder: "desc",
  current: { feedId: null, postFeedId: null, guid: null },
  dialog: { help: false },
});

/** The app's state. Read it in components; change it through app actions. */
export const app = state;

/** The store writer for actions in `~/stores/app`. */
export const setAppState = setState;

/** The selected subscription, if any. */
export function currentFeed(): Feed | undefined {
  return state.feeds.find((feed) => feed.id === state.current.feedId);
}

/** The title of a post's subscription, useful in the aggregate list. */
export function feedTitle(id: number): string {
  return state.feeds.find((feed) => feed.id === id)?.title ?? "";
}

/** Whether a feed is being fetched right now. */
export function isSyncingFeed(id: number): boolean {
  return state.sync.active && state.sync.feedId === id;
}
