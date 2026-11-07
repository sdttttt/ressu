import { createStore } from "solid-js/store";
import { toast } from "solid-sonner";

import { externalHref } from "~/lib/sanitize";
import { generateSnapshot } from "~/lib/snapshot";
import { applyFontSize, applyTheme } from "~/lib/theme";
import {
  Feeds,
  Posts,
  Settings,
  type Feed,
  type Post,
  type PostSummary,
  type Prefs,
  type SnapshotTile,
  type SyncState,
} from "~/mygo";

/** How many posts a page of the list holds. */
export const PAGE_SIZE = 100;

export type DialogName = "addFeed" | "settings" | "help";

type State = {
  feeds: Feed[];
  posts: PostSummary[];
  postTotal: number;
  postUnread: number;
  post: Post | null;
  dataDir: string;
  prefs: Prefs;
  sync: SyncState;
  syncErrors: string[];
  ready: boolean;
  loadingPosts: boolean;
  loadingPost: boolean;
  snapshotting: boolean;
  feedFilter: string;
  unreadOnly: boolean;
  starredOnly: boolean;
  current: { feedId: number | null; guid: string | null };
  dialog: Record<DialogName, boolean>;
};

const defaultPrefs: Prefs = {
  theme: "system",
  fontSize: "medium",
  refreshOnStart: true,
  markReadOnOpen: true,
  openLinks: "browser",
  postsPerFeed: 500,
  paneSizes: [18, 30],
};

const idleSync: SyncState = { active: false, done: 0, total: 0, feedId: 0, feed: "" };

const [state, setState] = createStore<State>({
  feeds: [],
  posts: [],
  postTotal: 0,
  postUnread: 0,
  post: null,
  dataDir: "",
  prefs: defaultPrefs,
  sync: { ...idleSync },
  syncErrors: [],
  ready: false,
  loadingPosts: false,
  loadingPost: false,
  snapshotting: false,
  feedFilter: "",
  unreadOnly: false,
  starredOnly: false,
  current: { feedId: null, guid: null },
  dialog: { addFeed: false, settings: false, help: false },
});

/** The app's state. Read it in components; change it through the actions. */
export const app = state;

/** The subscriptions matching the sidebar's search. */
export function visibleFeeds(): Feed[] {
  const needle = state.feedFilter.trim().toLowerCase();
  if (!needle) return state.feeds;
  return state.feeds.filter(
    (feed) => feed.title.toLowerCase().includes(needle) || feed.url.toLowerCase().includes(needle),
  );
}

/** The selected subscription, if any. */
export function currentFeed(): Feed | undefined {
  return state.feeds.find((feed) => feed.id === state.current.feedId);
}

/** How many posts are unread across the subscriptions. */
export function totalUnread(): number {
  return state.feeds.reduce((total, feed) => total + feed.unread, 0);
}

/** Whether a feed is being fetched right now. */
export function isSyncingFeed(id: number): boolean {
  return state.sync.active && state.sync.feedId === id;
}

/** The message of an error raised by a call into Go. */
export function messageOf(error: unknown): string {
  if (error instanceof Error) return error.message;
  return String(error);
}

export function openDialog(name: DialogName): void {
  setState("dialog", name, true);
}

export function closeDialog(name: DialogName): void {
  setState("dialog", name, false);
}

export function setFeedFilter(value: string): void {
  setState("feedFilter", value);
}

/** Loads the subscriptions and the preferences, and selects the first feed. */
export async function start(): Promise<void> {
  try {
    const [feeds, prefs, dataDir] = await Promise.all([
      Feeds.list(),
      Settings.get(),
      Settings.dataDir(),
    ]);
    setState({ feeds, prefs, dataDir, ready: true });
    applyTheme(prefs.theme);
    applyFontSize(prefs.fontSize);

    const first = feeds[0];
    if (first) await selectFeed(first.id);
  } catch (error) {
    setState("ready", true);
    toast.error(`无法读取本地数据：${messageOf(error)}`);
  }
}

/** Selects a subscription and loads the first page of its posts. */
export async function selectFeed(id: number): Promise<void> {
  setState("current", { feedId: id, guid: null });
  setState("post", null);
  await loadPosts(0);
}

/** Loads a page of the selected feed's posts, replacing or appending them. */
export async function loadPosts(offset = 0): Promise<void> {
  const feedId = state.current.feedId;
  if (feedId === null) {
    setState({ posts: [], postTotal: 0, postUnread: 0 });
    return;
  }

  setState("loadingPosts", true);
  try {
    const page = await Posts.list({
      feedId,
      unreadOnly: state.unreadOnly,
      starredOnly: state.starredOnly,
      search: "",
      offset,
      limit: PAGE_SIZE,
    });
    if (offset === 0) setState("posts", page.posts);
    else setState("posts", (posts) => [...posts, ...page.posts]);
    setState({ postTotal: page.total, postUnread: page.unread });
  } catch (error) {
    toast.error(`无法读取文章：${messageOf(error)}`);
  } finally {
    setState("loadingPosts", false);
  }
}

/** Loads the next page of posts. */
export async function loadMorePosts(): Promise<void> {
  if (state.loadingPosts) return;
  await loadPosts(state.posts.length);
}

/** Filters the list by unread or starred posts. */
export async function setPostFilter(filter: {
  unreadOnly?: boolean;
  starredOnly?: boolean;
}): Promise<void> {
  setState({
    unreadOnly: filter.unreadOnly ?? state.unreadOnly,
    starredOnly: filter.starredOnly ?? state.starredOnly,
  });
  await loadPosts(0);
}

/** Opens a post, and marks it read when the preferences ask for that. */
export async function selectPost(guid: string): Promise<void> {
  const feedId = state.current.feedId;
  if (feedId === null) return;

  setState("current", "guid", guid);
  setState("loadingPost", true);
  try {
    const post = await Posts.get(feedId, guid);
    setState("post", post);
    if (state.prefs.markReadOnOpen && !post.read) {
      await Posts.markRead(feedId, guid, true);
      setReadLocally(guid, true);
    }
  } catch (error) {
    toast.error(`无法打开文章：${messageOf(error)}`);
  } finally {
    setState("loadingPost", false);
  }
}

/** Marks the open post read or unread. */
export async function setOpenPostRead(read: boolean): Promise<void> {
  const feedId = state.current.feedId;
  const post = state.post;
  if (feedId === null || !post) return;

  try {
    await Posts.markRead(feedId, post.guid, read);
    setReadLocally(post.guid, read);
    setState("post", "read", read);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Marks every post of a feed read, the selected one by default. */
export async function markAllRead(feedId: number | null = state.current.feedId): Promise<void> {
  if (feedId === null) return;

  try {
    const changed = await Posts.markAllRead(feedId);
    if (changed === 0) {
      toast.info("这个订阅源没有未读文章");
      return;
    }
    toast.success(`已把 ${changed} 篇文章标记为已读`);
    if (feedId === state.current.feedId) await loadPosts(0);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Fetches one feed. */
export async function refreshFeed(feedId: number): Promise<void> {
  try {
    const feed = await Feeds.refresh(feedId);
    setState("feeds", (item) => item.id === feed.id, feed);
  } catch (error) {
    toast.error(`${feedId}：${messageOf(error)}`);
  }
}

/** Stars or unstars a post. */
export async function toggleStar(guid: string): Promise<void> {
  const feedId = state.current.feedId;
  if (feedId === null) return;

  try {
    const updated = await Posts.toggleStar(feedId, guid);
    setState("posts", (post) => post.guid === guid, "starred", updated.starred);
    if (state.post?.guid === guid) setState("post", "starred", updated.starred);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Captures the open post's tiles, holding `snapshotting` while it runs.
 * Returns null when there is nothing to do or generation failed; the post is
 * captured with the tiles so a switched article cannot mix them. */
async function withSnapshot(): Promise<{ post: Post; tiles: SnapshotTile[] } | null> {
  const post = state.post;
  if (!post || state.snapshotting) return null;
  const feedTitle = currentFeed()?.title ?? "";

  setState("snapshotting", true);
  try {
    const tiles = await generateSnapshot(post, feedTitle);
    return { post, tiles };
  } catch (error) {
    toast.error(`无法生成快照：${messageOf(error)}`);
    return null;
  } finally {
    setState("snapshotting", false);
  }
}

/** Saves the open post's snapshot as a PNG through a save dialog. */
export async function saveSnapshot(): Promise<void> {
  const snapshot = await withSnapshot();
  if (!snapshot) return;
  try {
    const path = await Posts.saveSnapshot(snapshot.post.title, snapshot.tiles);
    if (!path) return; // The dialog was cancelled; stay quiet.
    toast.success(`快照已保存：${path}`);
  } catch (error) {
    toast.error(`无法保存快照：${messageOf(error)}`);
  }
}

/** Puts the open post's snapshot on the clipboard as a PNG. */
export async function copySnapshot(): Promise<void> {
  const snapshot = await withSnapshot();
  if (!snapshot) return;
  try {
    await Posts.copySnapshot(snapshot.tiles);
    toast.success("快照已复制到剪贴板");
  } catch (error) {
    toast.error(`无法复制快照：${messageOf(error)}`);
  }
}

/** Refreshes every subscription. */
export async function refreshAll(): Promise<void> {
  if (state.sync.active) return;

  try {
    const summary = await Feeds.refreshAll();
    if (summary.updated + summary.failed === 0) {
      toast.info("还没有订阅源");
    } else if (summary.failed > 0) {
      toast.warning(`${summary.updated} 个订阅源已更新，${summary.failed} 个失败`);
    } else if (summary.new > 0) {
      toast.success(`新增 ${summary.new} 篇文章`);
    } else {
      toast.success("所有订阅源都已是最新");
    }
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Subscribes to a feed. Returns whether it worked. */
export async function addFeed(feedURL: string): Promise<boolean> {
  try {
    const feed = await Feeds.add(feedURL.trim());
    setState("feeds", (feeds) => [...feeds.filter((item) => item.id !== feed.id), feed]);
    closeDialog("addFeed");
    toast.success(`已添加「${feed.title}」`);
    await selectFeed(feed.id);
    return true;
  } catch (error) {
    toast.error(messageOf(error));
    return false;
  }
}

/** Unsubscribes from a feed, selecting another one. */
export async function removeFeed(id: number): Promise<void> {
  const wasCurrent = state.current.feedId === id;
  try {
    await Feeds.remove(id);
    const feeds = state.feeds.filter((feed) => feed.id !== id);
    setState("feeds", feeds);
    if (!wasCurrent) return;

    const next = feeds[0];
    if (next) {
      await selectFeed(next.id);
    } else {
      setState({
        posts: [],
        postTotal: 0,
        postUnread: 0,
        post: null,
        current: { feedId: null, guid: null },
      });
    }
    toast.success("已取消订阅");
  } catch (error) {
    toast.error(messageOf(error));
  }
}

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

/** Applies the feed list the Go side sent. */
export function replaceFeeds(feeds: Feed[]): void {
  setState("feeds", feeds);
  if (state.current.feedId !== null && !feeds.some((feed) => feed.id === state.current.feedId)) {
    const next = feeds[0];
    if (next) void selectFeed(next.id);
    else
      setState({
        posts: [],
        postTotal: 0,
        postUnread: 0,
        post: null,
        current: { feedId: null, guid: null },
      });
  }
}

/** Applies a refresh event. */
export function applySync(event: SyncState): void {
  if (event.active) {
    setState("sync", event);
    if (event.error) {
      setState("syncErrors", (errors) => [...errors, `${event.feed || "订阅源"}：${event.error}`]);
    }
    return;
  }

  setState("sync", { ...idleSync, done: event.total, total: event.total, error: undefined });
  if (state.syncErrors.length > 0) {
    toast.error(`${state.syncErrors.length} 个订阅源刷新失败，详情见状态栏`);
    setState("syncErrors", []);
  }
}

/** Refreshes the list of the selected feed after its posts changed. */
export function handlePostsChanged(feedId: number): void {
  if (feedId === state.current.feedId) void loadPosts(0);
}

function setReadLocally(guid: string, read: boolean): void {
  setState("posts", (post) => post.guid === guid, "read", read);
  setState("postUnread", (unread) => Math.max(read ? unread - 1 : unread + 1, 0));
}

/** The list index the selection would move to, honouring the unread-only
 * filter. Returns -1 when there is nothing to move onto. */
function selectionTarget(delta: 1 | -1, onlyUnread: boolean): number {
  const posts = state.posts;
  if (posts.length === 0) return -1;

  const current = posts.findIndex((post) => post.guid === state.current.guid);
  let index = current === -1 ? (delta === 1 ? 0 : posts.length - 1) : current + delta;

  while (index >= 0 && index < posts.length) {
    if (!onlyUnread || !posts[index]!.read) return index;
    index += delta;
  }
  return -1;
}

/** Scrolls the post list so the row of `guid` is visible. */
function revealPost(guid: string): void {
  document
    .querySelector<HTMLElement>(`[data-guid="${CSS.escape(guid)}"]`)
    ?.scrollIntoView({ block: "nearest" });
}

/**
 * Moves the selection within the loaded posts, optionally skipping read ones,
 * and opens the post it lands on. Stops at the ends of the list instead of
 * wrapping around.
 */
export async function moveSelection(delta: 1 | -1, onlyUnread = false): Promise<void> {
  const index = selectionTarget(delta, onlyUnread);
  if (index === -1) return;

  const post = state.posts[index]!;
  await selectPost(post.guid);
  revealPost(post.guid);
}

/** Opens the selected post, or the first post when nothing is selected. */
export async function openSelectedPost(): Promise<void> {
  const index = selectionTarget(1, false);
  if (index === -1) return;

  const post = state.posts[index]!;
  await selectPost(post.guid);
  revealPost(post.guid);
}

/** Opens the selected post's original page in a browser. */
export async function openSelectedPostInBrowser(): Promise<void> {
  const post = state.posts.find((item) => item.guid === state.current.guid);
  if (post) await openLink(post.url);
}

/** Toggles the read state of the open post. */
export async function toggleOpenPostRead(): Promise<void> {
  const post = state.post;
  if (post) await setOpenPostRead(!post.read);
}

/** Selects the feed next to or before the selected one. */
export async function selectNeighbourFeed(delta: 1 | -1): Promise<void> {
  const feeds = state.feeds;
  if (feeds.length === 0) return;

  const current = feeds.findIndex((feed) => feed.id === state.current.feedId);
  const index = current === -1 ? (delta === 1 ? 0 : feeds.length - 1) : current + delta;
  const feed = feeds[index];
  if (feed) await selectFeed(feed.id);
}
