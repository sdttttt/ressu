import { beforeEach, describe, expect, it, vi } from "vitest";

import type {
  Event as BroadcastEvent,
  Feed,
  FeedsSnapshot,
  Post,
  PostSummary,
  Prefs,
  SyncState,
} from "~/mygo";

vi.mock("~/lib/notifications", () => ({
  toast: { success: vi.fn(), error: vi.fn(), info: vi.fn(), warning: vi.fn() },
}));

vi.mock("~/lib/snapshot", () => ({ generateSnapshot: vi.fn() }));

vi.mock("~/mygo", () => ({
  Feeds: {
    list: vi.fn(),
    openAddFeed: vi.fn(),
    refreshAll: vi.fn(),
  },
  Posts: {
    list: vi.fn(),
    get: vi.fn(),
    markRead: vi.fn(),
    markAllRead: vi.fn(),
    toggleStar: vi.fn(),
    captureSnapshot: vi.fn(),
    stitchSnapshot: vi.fn(),
    saveSnapshot: vi.fn(),
    copySnapshot: vi.fn(),
  },
  Settings: {
    get: vi.fn(),
    update: vi.fn(),
    openLink: vi.fn(),
    openSettings: vi.fn(),
  },
  Status: {
    selectFeed: vi.fn(),
    pageReady: vi.fn(),
    notify: vi.fn(),
  },
  events: {
    appEvent: { on: vi.fn(() => () => {}) },
  },
}));

const feedA: Feed = {
  id: 1,
  url: "https://a.example/feed",
  title: "A 站",
  siteUrl: "https://a.example/",
  description: "",
  iconUrl: "",
  language: "zh",
  etag: "",
  lastModified: "",
  lastFetchedAt: null,
  lastError: "",
  unread: 1,
  createdAt: "2026-01-01T00:00:00Z",
};

const feedB: Feed = { ...feedA, id: 2, url: "https://b.example/feed", title: "B 站", unread: 0 };

const summary: PostSummary = {
  feedId: 1,
  guid: "guid-1",
  title: "第一篇文章",
  url: "https://a.example/1",
  author: "",
  summary: "摘要",
  wordCount: 3,
  publishedAt: "2026-01-02T00:00:00Z",
  read: false,
  starred: false,
};

const post: Post = {
  ...summary,
  content: "<p>正文</p>",
  wordCount: 3,
  fetchedAt: "2026-01-02T00:00:00Z",
};

const twoPosts = (): PostSummary[] => [
  summary,
  { ...summary, guid: "guid-2", title: "第二篇文章", read: true },
  { ...summary, guid: "guid-3", title: "第三篇文章" },
];

const prefs: Prefs = {
  theme: "system",
  fontSize: "small",
  refreshOnStart: true,
  markReadOnOpen: true,
  remoteResources: "goProxy",
  openLinks: "browser",
  postsPerFeed: 500,
  imageCacheLimit: 256,
  paneSizes: [18, 30],
  sidebarCollapsed: false,
};

beforeEach(() => {
  vi.resetModules();
  vi.clearAllMocks();
});

async function setup(options: { feeds?: Feed[]; unread?: number; posts?: PostSummary[] } = {}) {
  const mygo = await import("~/mygo");
  const snapshot = await import("~/lib/snapshot");
  const { toast } = await import("~/lib/notifications");
  const app = await import("~/stores/app");
  const storeFeeds = await import("~/stores/feeds");
  const syncEvents = await import("~/stores/sync-events");

  const feeds = (options.feeds ?? [feedA, feedB]).map((feed) => ({ ...feed }));
  const feedsSnapshot: FeedsSnapshot = {
    feeds,
    unread: options.unread ?? feeds.reduce((total, feed) => total + feed.unread, 0),
  };
  vi.mocked(mygo.Feeds.list).mockResolvedValue(feedsSnapshot);
  vi.mocked(mygo.Settings.get).mockResolvedValue({ ...prefs });
  vi.mocked(mygo.Settings.update).mockImplementation(async (next) => next);
  vi.mocked(mygo.Status.pageReady).mockResolvedValue(undefined);
  vi.mocked(mygo.Status.selectFeed).mockResolvedValue(undefined);
  vi.mocked(mygo.Posts.list).mockImplementation(async () => {
    const posts = (options.posts ?? [summary]).map((item) => ({ ...item }));
    return { posts, total: posts.length, unread: posts.filter((item) => !item.read).length };
  });
  vi.mocked(mygo.Posts.get).mockImplementation(async (_feedId: number, guid: string) => ({
    ...post,
    guid,
  }));
  vi.mocked(mygo.Posts.markRead).mockResolvedValue(undefined);

  return { mygo, snapshot, toast, app, storeFeeds, syncEvents };
}

function makeEvent(
  eid: BroadcastEvent["eid"],
  overrides: Partial<BroadcastEvent> = {},
): BroadcastEvent {
  return {
    eid,
    revision: 1,
    occurredAt: "2026-11-09T12:00:00Z",
    sync: null,
    feeds: null,
    posts: null,
    appearance: null,
    prefs: null,
    selection: null,
    postSelection: null,
    post: null,
    command: "",
    ...overrides,
  };
}

describe("sidebar preference", () => {
  it("loads every feed's posts without selecting one", async () => {
    const { mygo, app } = await setup();

    await app.start();

    expect(app.app.ready).toBe(true);
    expect(app.app.feeds).toHaveLength(2);
    expect(app.app.totalUnread).toBe(1);
    expect(app.app.current.feedId).toBeNull();
    expect(app.app.posts).toHaveLength(1);
    expect(mygo.Posts.list).toHaveBeenLastCalledWith({
      feedId: 0,
      unreadOnly: false,
      starredOnly: false,
      search: "",
      offset: 0,
      limit: 100,
      sortOrder: "desc",
    });
  });

  it("survives a failing store", async () => {
    const { mygo, app } = await setup();
    vi.mocked(mygo.Feeds.list).mockRejectedValue(new Error("boom"));

    await app.start();

    expect(app.app.ready).toBe(true);
    expect(app.app.feeds).toHaveLength(0);
  });
});

describe("selectPost", () => {
  it("opens the post and marks it read", async () => {
    const { mygo, app } = await setup();
    await app.start();

    await app.selectPost("guid-1");

    expect(app.app.post?.guid).toBe("guid-1");
    expect(mygo.Posts.markRead).toHaveBeenCalledWith(1, "guid-1", true);
    expect(app.app.posts[0]?.read).toBe(true);
    expect(app.app.postUnread).toBe(0);
  });

  it("does not decrement unread twice when the update event arrives first", async () => {
    const { app, syncEvents } = await setup();
    await app.start();

    syncEvents.handlePostUpdated({ ...summary, read: true });
    await app.selectPost("guid-1");

    expect(app.app.posts[0]?.read).toBe(true);
    expect(app.app.postUnread).toBe(0);
  });

  it("leaves the post unread when the preferences say so", async () => {
    const { mygo, app } = await setup();
    vi.mocked(mygo.Settings.get).mockResolvedValue({ ...prefs, markReadOnOpen: false });
    await app.start();

    await app.selectPost("guid-1");

    expect(mygo.Posts.markRead).not.toHaveBeenCalled();
    expect(app.app.posts[0]?.read).toBe(false);
  });
});

describe("selectFeed", () => {
  it("loads a selected feed and toggles it back to aggregate", async () => {
    const { mygo, app, storeFeeds } = await setup();
    await app.start();

    await storeFeeds.selectFeed(2);

    expect(app.app.current.feedId).toBe(2);
    expect(mygo.Posts.list).toHaveBeenLastCalledWith({
      feedId: 2,
      unreadOnly: false,
      starredOnly: false,
      search: "",
      offset: 0,
      limit: 100,
      sortOrder: "desc",
    });

    await storeFeeds.selectFeed(2);

    expect(app.app.current.feedId).toBeNull();
    expect(mygo.Posts.list).toHaveBeenLastCalledWith({
      feedId: 0,
      unreadOnly: false,
      starredOnly: false,
      search: "",
      offset: 0,
      limit: 100,
      sortOrder: "desc",
    });
  });
});

describe("setPostFilter", () => {
  it("asks Go for the filtered posts", async () => {
    const { mygo, app } = await setup();
    await app.start();

    await app.setPostFilter({ unreadOnly: true });

    expect(app.app.unreadOnly).toBe(true);
    expect(vi.mocked(mygo.Posts.list).mock.lastCall?.[0]).toMatchObject({ unreadOnly: true });
  });
});

describe("post search", () => {
  it("searches titles after typing settles", async () => {
    vi.useFakeTimers();
    const { mygo, app } = await setup();
    await app.start();

    app.setPostSearch("标题");
    expect(app.app.postSearch).toBe("");
    expect(vi.mocked(mygo.Posts.list)).not.toHaveBeenCalledWith(
      expect.objectContaining({ search: "标题" }),
    );

    vi.advanceTimersByTime(300);
    expect(app.app.postSearch).toBe("标题");
    expect(vi.mocked(mygo.Posts.list).mock.lastCall?.[0]).toMatchObject({
      offset: 0,
      search: "标题",
    });
    vi.useRealTimers();
  });

  it("clears the query when the search field is closed", async () => {
    const { mygo, app } = await setup();
    await app.start();
    app.setPostSearch("标题");
    await new Promise((resolve) => setTimeout(resolve, 320));

    await app.togglePostSearch();
    await app.togglePostSearch();

    expect(app.app.postSearchOpen).toBe(false);
    expect(app.app.postSearch).toBe("");
    expect(vi.mocked(mygo.Posts.list).mock.lastCall?.[0]).toMatchObject({
      offset: 0,
      search: "",
    });
  });
});

describe("setPostSortOrder", () => {
  it("asks Go for the sorted first page", async () => {
    const { mygo, app } = await setup();
    await app.start();

    await app.setPostSortOrder("asc");

    expect(app.app.postSortOrder).toBe("asc");
    expect(vi.mocked(mygo.Posts.list).mock.lastCall?.[0]).toMatchObject({
      offset: 0,
      sortOrder: "asc",
    });
  });
});

describe("openAddFeed", () => {
  it("opens the native window", async () => {
    const { mygo, app } = await setup();
    await app.start();
    vi.mocked(mygo.Feeds.openAddFeed).mockResolvedValue(undefined);

    await app.openAddFeed();

    expect(mygo.Feeds.openAddFeed).toHaveBeenCalled();
  });

  it("reports an open failure", async () => {
    const { mygo, app } = await setup();
    await app.start();
    const { toast } = await import("~/lib/notifications");
    vi.mocked(mygo.Feeds.openAddFeed).mockRejectedValue(new Error("窗口不可用"));

    await app.openAddFeed();

    expect(toast.error).toHaveBeenCalledWith("窗口不可用");
  });
});

describe("applySync", () => {
  it("updates the page when Go broadcasts a sync state", async () => {
    const { mygo, app } = await setup();
    const { subscribe } = await import("~/stores/events");
    subscribe();
    await app.start();

    const listener = vi.mocked(mygo.events.appEvent.on).mock.calls[0]?.[0];
    expect(listener).toBeTypeOf("function");
    listener!(
      makeEvent("sync.progress", {
        sync: { active: true, done: 1, total: 2, feedId: 1, feed: "A 站" },
      }),
    );

    expect(app.app.sync).toMatchObject({ active: true, done: 1, total: 2, feed: "A 站" });
  });

  it("ignores an event with an older revision", async () => {
    const { mygo, app } = await setup();
    const { subscribe } = await import("~/stores/events");
    subscribe();
    await app.start();

    const listener = vi.mocked(mygo.events.appEvent.on).mock.calls[0]?.[0];
    listener!(
      makeEvent("sync.progress", {
        revision: 3,
        sync: { active: true, done: 2, total: 2, feedId: 1, feed: "A 站" },
      }),
    );
    listener!(
      makeEvent("sync.progress", {
        revision: 2,
        sync: { active: true, done: 1, total: 2, feedId: 1, feed: "A 站" },
      }),
    );

    expect(app.app.sync.done).toBe(2);
  });

  it("applies a full state snapshot", async () => {
    const { mygo, app } = await setup();
    const { subscribe } = await import("~/stores/events");
    subscribe();
    await app.start();

    const listener = vi.mocked(mygo.events.appEvent.on).mock.calls[0]?.[0];
    listener!(
      makeEvent("state.snapshot", {
        revision: 2,
        feeds: { feeds: [feedB], unread: 5 },
        sync: { active: false, done: 0, total: 0, feedId: 0, feed: "" },
        appearance: { theme: "dark", fontSize: "large" },
        selection: { feedId: 0 },
      }),
    );

    await vi.waitFor(() => expect(app.app.feeds).toHaveLength(1));
    expect(app.app.totalUnread).toBe(5);
    expect(app.app.prefs.theme).toBe("dark");
    expect(app.app.prefs.fontSize).toBe("large");
  });

  it("applies preferences changed by native settings", async () => {
    const { mygo, app } = await setup();
    const { subscribe } = await import("~/stores/events");
    subscribe();
    await app.start();

    const listener = vi.mocked(mygo.events.appEvent.on).mock.calls[0]?.[0];
    listener!(
      makeEvent("prefs.changed", {
        revision: 2,
        prefs: { ...prefs, theme: "sepia", fontSize: "large" },
      }),
    );

    await vi.waitFor(() => expect(app.app.prefs.theme).toBe("sepia"));
    expect(app.app.prefs.fontSize).toBe("large");
    expect(document.documentElement.classList.contains("theme-sepia")).toBe(true);
  });

  it("ends the sync state after the last feed", async () => {
    const { app, syncEvents } = await setup();
    await app.start();

    const active = (done: number, error?: string): SyncState => ({
      active: true,
      done,
      total: 2,
      feedId: 1,
      feed: "A 站",
      ...(error ? { error } : {}),
    });
    syncEvents.applySync(active(0));
    syncEvents.applySync(active(1, "服务器错误"));
    expect(app.app.sync.active).toBe(true);
    expect(app.app.syncErrors).toEqual(["A 站：服务器错误"]);

    syncEvents.applySync({ active: false, done: 2, total: 2, feedId: 0, feed: "" });
    expect(app.app.sync.active).toBe(false);
    expect(app.app.sync.done).toBe(2);
    expect(app.app.sync.total).toBe(2);
    expect(app.app.syncErrors).toEqual([]);
  });

  it("resets a sync that is already active", async () => {
    const { app, syncEvents } = await setup();
    await app.start();

    syncEvents.applySync({ active: true, done: 0, total: 2, feedId: 1, feed: "A 站" });
    // A second refresh cannot run while the first one does, so the state the
    // page holds must recover when Go announces a new run.
    syncEvents.applySync({ active: true, done: 0, total: 2, feedId: 1, feed: "A 站" });
    syncEvents.applySync({ active: false, done: 2, total: 2, feedId: 0, feed: "" });
    expect(app.app.sync.active).toBe(false);
  });
});

describe("replaceFeeds", () => {
  it("stores backend unread and returns to aggregate when its feed disappears", async () => {
    const { app, syncEvents } = await setup();
    await app.start();

    await syncEvents.replaceFeeds({ feeds: [feedB], unread: 5 });
    await vi.waitFor(() => expect(app.app.current.feedId).toBeNull());
    expect(app.app.feeds).toHaveLength(1);
    expect(app.app.totalUnread).toBe(5);
    expect(app.app.posts).toHaveLength(1);
  });
});

describe("moveSelection", () => {
  it("moves the selection down and opens the post", async () => {
    const { mygo, app } = await setup({ posts: twoPosts() });
    await app.start();
    await app.selectPost("guid-1");

    await app.moveSelection(1);

    expect(app.app.current.guid).toBe("guid-2");
    expect(app.app.post?.guid).toBe("guid-2");
    expect(mygo.Posts.get).toHaveBeenLastCalledWith(1, "guid-2");
  });

  it("stops at the ends of the list instead of wrapping", async () => {
    const { app } = await setup({ posts: twoPosts() });
    await app.start();
    await app.selectPost("guid-1");

    await app.moveSelection(-1);

    expect(app.app.current.guid).toBe("guid-1");

    await app.selectPost("guid-3");
    await app.moveSelection(1);

    expect(app.app.current.guid).toBe("guid-3");
  });

  it("skips read posts when only unread ones are wanted", async () => {
    const { mygo, app } = await setup({ posts: twoPosts() });
    vi.mocked(mygo.Settings.get).mockResolvedValue({ ...prefs, markReadOnOpen: false });
    await app.start();
    await app.selectPost("guid-1");

    await app.moveSelection(1, true);

    expect(app.app.current.guid).toBe("guid-3");

    await app.moveSelection(-1, true);

    expect(app.app.current.guid).toBe("guid-1");
  });

  it("does nothing when there are no posts", async () => {
    const { mygo, app } = await setup({ posts: [] });
    await app.start();

    await app.moveSelection(1);

    expect(mygo.Posts.get).not.toHaveBeenCalled();
    expect(app.app.current.guid).toBeNull();
  });

  it("starts at the first post when nothing is selected yet", async () => {
    const { app } = await setup({ posts: twoPosts() });
    await app.start();
    app.app.current.guid = null;

    await app.openSelectedPost();

    expect(app.app.current.guid).toBe("guid-1");
    expect(app.app.post?.guid).toBe("guid-1");
  });
});

describe("toggleOpenPostRead", () => {
  it("flips the read state of the open post", async () => {
    const { mygo, app } = await setup();
    vi.mocked(mygo.Settings.get).mockResolvedValue({ ...prefs, markReadOnOpen: false });
    await app.start();
    await app.selectPost("guid-1");

    await app.toggleOpenPostRead();

    expect(mygo.Posts.markRead).toHaveBeenCalledWith(1, "guid-1", true);
    expect(app.app.post?.read).toBe(true);

    await app.toggleOpenPostRead();

    expect(mygo.Posts.markRead).toHaveBeenLastCalledWith(1, "guid-1", false);
  });

  it("does nothing when no post is open", async () => {
    const { mygo, app } = await setup({ posts: [] });
    await app.start();

    await app.toggleOpenPostRead();

    expect(mygo.Posts.markRead).not.toHaveBeenCalled();
  });
});

describe("openSelectedPostInBrowser", () => {
  it("opens the selected post's URL through Settings", async () => {
    const { mygo, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");

    await app.openSelectedPostInBrowser();

    expect(mygo.Settings.openLink).toHaveBeenCalledWith("https://a.example/1");
  });
});

describe("selectNeighbourFeed", () => {
  it("walks between neighbouring feeds", async () => {
    const { app } = await setup();
    await app.start();

    await app.selectNeighbourFeed(1);
    expect(app.app.current.feedId).toBe(1);

    await app.selectNeighbourFeed(1);
    expect(app.app.current.feedId).toBe(2);

    await app.selectNeighbourFeed(-1);
    expect(app.app.current.feedId).toBe(1);
  });

  it("stops at the neighbouring feed boundaries", async () => {
    const { app } = await setup();
    await app.start();

    await app.selectNeighbourFeed(-1);
    expect(app.app.current.feedId).toBe(2);

    await app.selectNeighbourFeed(1);
    expect(app.app.current.feedId).toBe(2);

    await app.selectNeighbourFeed(-1);
    expect(app.app.current.feedId).toBe(1);
  });

  it("stops at the ends of the feed list", async () => {
    const { app } = await setup({ feeds: [feedA] });
    await app.start();

    await app.selectNeighbourFeed(1);

    expect(app.app.current.feedId).toBe(1);
  });
});

describe("saveSnapshot", () => {
  it("stitches the open post and opens its preview", async () => {
    const { mygo, snapshot, toast, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    const tiles = [
      {
        offset: 0,
        png: "cG5n",
        geometry: { viewportWidth: 1000, viewportHeight: 800, cardWidth: 760, contentHeight: 1600 },
      },
    ];
    vi.mocked(snapshot.generateSnapshot).mockResolvedValue(tiles);
    vi.mocked(mygo.Posts.stitchSnapshot).mockResolvedValue("c3RpdGNoZWQ");

    await app.captureSnapshot();

    expect(snapshot.generateSnapshot).toHaveBeenCalledWith(
      expect.objectContaining({ guid: "guid-1" }),
      "A 站",
    );
    expect(mygo.Posts.stitchSnapshot).toHaveBeenCalledWith(tiles);
    expect(app.app.snapshot).toEqual({
      post: expect.objectContaining({ guid: "guid-1" }),
      png: "c3RpdGNoZWQ",
    });
    expect(app.app.snapshotting).toBe(false);
    expect(toast.error).not.toHaveBeenCalled();
  });

  it("reports stitching failures without opening a preview", async () => {
    const { mygo, snapshot, toast, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    vi.mocked(snapshot.generateSnapshot).mockResolvedValue([]);
    vi.mocked(mygo.Posts.stitchSnapshot).mockRejectedValue(new Error("bad tiles"));

    await app.captureSnapshot();

    expect(app.app.snapshot).toBeNull();
    expect(app.app.snapshotting).toBe(false);
    expect(toast.error).toHaveBeenCalledWith("无法生成快照：bad tiles");
  });

  it("does not start while another snapshot is running", async () => {
    const { snapshot, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    let release: (() => void) | undefined;
    vi.mocked(snapshot.generateSnapshot).mockImplementation(
      () =>
        new Promise((resolve) => {
          release = () => resolve([]);
        }),
    );

    const first = app.captureSnapshot();
    await app.captureSnapshot();
    expect(snapshot.generateSnapshot).toHaveBeenCalledTimes(1);
    release?.();
    await first;
  });

  async function preview(app: Awaited<ReturnType<typeof setup>>["app"]) {
    const mygo = await import("~/mygo");
    const snapshot = await import("~/lib/snapshot");
    vi.mocked(snapshot.generateSnapshot).mockResolvedValue([]);
    vi.mocked(mygo.Posts.stitchSnapshot).mockResolvedValue("c3RpdGNoZWQ");
    await app.captureSnapshot();
  }

  it("hands the previewed PNG to Go and closes the preview", async () => {
    const { mygo, toast, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    await preview(app);
    vi.mocked(mygo.Posts.saveSnapshot).mockResolvedValue("C:/Users/me/第一篇文章.png");

    await app.saveSnapshot();

    expect(mygo.Posts.saveSnapshot).toHaveBeenCalledWith("第一篇文章", "c3RpdGNoZWQ");
    expect(toast.success).not.toHaveBeenCalled();
    expect(app.app.snapshot).toBeNull();
    expect(app.app.snapshotAction).toBeNull();
  });

  it("keeps the preview open when the save dialog is cancelled", async () => {
    const { mygo, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    await preview(app);
    vi.mocked(mygo.Posts.saveSnapshot).mockResolvedValue("");

    await app.saveSnapshot();

    expect(mygo.Posts.saveSnapshot).toHaveBeenCalledWith("第一篇文章", "c3RpdGNoZWQ");
    expect(app.app.snapshot).not.toBeNull();
    expect(app.app.snapshotAction).toBeNull();
  });

  it("reports a failing save and keeps the preview", async () => {
    const { mygo, toast, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    await preview(app);
    vi.mocked(mygo.Posts.saveSnapshot).mockRejectedValue(new Error("disk full"));

    await app.saveSnapshot();

    expect(toast.error).toHaveBeenCalledWith("无法保存快照：disk full");
    expect(app.app.snapshot).not.toBeNull();
    expect(app.app.snapshotAction).toBeNull();
  });

  it("does nothing without a preview", async () => {
    const { mygo, snapshot, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");

    await app.saveSnapshot();

    expect(snapshot.generateSnapshot).not.toHaveBeenCalled();
    expect(mygo.Posts.saveSnapshot).not.toHaveBeenCalled();
  });
});

describe("copySnapshot", () => {
  it("copies the previewed PNG through Go and closes the preview", async () => {
    const { mygo, snapshot, toast, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    vi.mocked(snapshot.generateSnapshot).mockResolvedValue([]);
    vi.mocked(mygo.Posts.stitchSnapshot).mockResolvedValue("c3RpdGNoZWQ");
    await app.captureSnapshot();
    vi.mocked(mygo.Posts.copySnapshot).mockResolvedValue(undefined);

    await app.copySnapshot();

    expect(mygo.Posts.copySnapshot).toHaveBeenCalledWith("c3RpdGNoZWQ");
    expect(toast.success).not.toHaveBeenCalled();
    expect(app.app.snapshot).toBeNull();
  });

  it("reports generation failures instead of calling Go", async () => {
    const { mygo, snapshot, toast, app } = await setup();
    await app.start();
    await app.selectPost("guid-1");
    vi.mocked(snapshot.generateSnapshot).mockRejectedValue(new Error("文章太长，无法生成一张快照"));

    await app.captureSnapshot();

    expect(mygo.Posts.copySnapshot).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith("无法生成快照：文章太长，无法生成一张快照");
    expect(app.app.snapshotting).toBe(false);
  });
});
