import { fireEvent, render } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";

import PostList from "~/components/post-list";

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

  const first = {
    feedId: 1,
    guid: "guid-1",
    title: "第一篇文章",
    url: "https://a.example/1",
    author: "作者甲",
    summary: "这是摘要",
    publishedAt: "2026-01-02T00:00:00Z",
    read: false,
    starred: false,
  };

  const second = {
    ...first,
    guid: "guid-2",
    title: "第二篇文章",
    summary: "",
    publishedAt: "2025-06-01T00:00:00Z",
    read: true,
    starred: true,
  };

  return { feed, first, second };
});

vi.mock("~/stores/app", () => ({
  app: {
    posts: [fixtures.first, fixtures.second],
    postTotal: 3,
    postUnread: 1,
    unreadOnly: false,
    starredOnly: false,
    loadingPosts: false,
    current: { feedId: 1, guid: null },
  },
  currentFeed: () => fixtures.feed,
  selectPost: vi.fn(),
  toggleStar: vi.fn(),
  setPostFilter: vi.fn(),
  markAllRead: vi.fn(),
  loadMorePosts: vi.fn(),
}));

describe("PostList", () => {
  it("shows the title, the summary and the date of each post", () => {
    const { getByText } = render(() => <PostList />);

    expect(getByText("第一篇文章")).toBeTruthy();
    expect(getByText("这是摘要")).toBeTruthy();
    expect(getByText(/2025/)).toBeTruthy();
    expect(getByText("A 站")).toBeTruthy();
  });

  it("opens a post when its row is clicked", async () => {
    const store = await import("~/stores/app");
    const { getByText } = render(() => <PostList />);

    fireEvent.click(getByText("第一篇文章"));

    expect(store.selectPost).toHaveBeenCalledWith("guid-1");
  });

  it("stars a post without opening it", async () => {
    const store = await import("~/stores/app");
    const { getAllByTitle } = render(() => <PostList />);

    const [star] = getAllByTitle("加星标");
    fireEvent.click(star!);

    expect(store.toggleStar).toHaveBeenCalledWith("guid-1");
    expect(store.selectPost).not.toHaveBeenCalled();
  });

  it("offers to load more posts", async () => {
    const store = await import("~/stores/app");
    const { getByText } = render(() => <PostList />);

    fireEvent.click(getByText(/加载更多/));

    expect(store.loadMorePosts).toHaveBeenCalled();
  });
});
