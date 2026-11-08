import { render, waitFor } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";

import FeedSidebar from "~/components/feed-sidebar";

const fixtures = vi.hoisted(() => {
  const feed = {
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
    unread: 2,
    createdAt: "2026-01-01T00:00:00Z",
  };
  const actions = {
    collapseSidebar: vi.fn(),
    expandSidebar: vi.fn(),
    ensureFeedIcon: vi.fn(),
    isSyncingFeed: vi.fn(() => false),
    markAllRead: vi.fn(),
    openDialog: vi.fn(),
    refreshAll: vi.fn(),
    refreshFeed: vi.fn(),
    removeFeed: vi.fn(),
    selectFeed: vi.fn(),
    setFeedFilter: vi.fn(),
  };
  return { feed, actions };
});

vi.mock("~/stores/app", () => ({
  app: {
    feeds: [fixtures.feed],
    icons: {
      [fixtures.feed.id]: {
        src: "data:image/png;base64,png-bytes",
        sourceURL: "https://a.example/favicon.ico",
        fetchedAt: "2026-01-01T00:00:00Z",
        missing: false,
      },
    },
    feedFilter: "",
    sidebar: { collapsed: false, visible: false },
    sync: { active: false },
    current: { feedId: fixtures.feed.id, postFeedId: null, guid: null },
  },
  currentFeed: () => fixtures.feed,
  visibleFeeds: () => [fixtures.feed],
  ...fixtures.actions,
}));

describe("FeedSidebar", () => {
  it("renders a local icon without deriving a remote favicon", async () => {
    const { container } = render(() => <FeedSidebar />);

    await waitFor(() => expect(fixtures.actions.ensureFeedIcon).toHaveBeenCalledWith(1));
    const image = container.querySelector("img");
    expect(image).not.toBeNull();
    expect((image as HTMLImageElement).src).toBe("data:image/png;base64,png-bytes");
  });
});
