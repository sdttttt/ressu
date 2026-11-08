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
    wordCount: 250,
    publishedAt: "2026-01-02T00:00:00Z",
    read: false,
    starred: false,
  };

  const second = {
    ...first,
    guid: "guid-2",
    title: "第二篇文章",
    summary: "",
    wordCount: 0,
    publishedAt: "2025-06-01T00:00:00Z",
    read: true,
    starred: true,
  };

  return { feed, first, second };
});

const { setPostSearch, setPostSortOrder, storeApp, togglePostSearch } = vi.hoisted(() => ({
  setPostSearch: vi.fn(),
  setPostSortOrder: vi.fn(),
  storeApp: {
    posts: [fixtures.first, fixtures.second],
    postTotal: 3,
    postUnread: 1,
    unreadOnly: false,
    starredOnly: false,
    postSearch: "",
    postSearchOpen: false,
    postSortOrder: "desc",
    loadingPosts: false,
    current: { feedId: 1, postFeedId: null, guid: null },
  },
  togglePostSearch: vi.fn(() => {
    storeApp.postSearchOpen = !storeApp.postSearchOpen;
  }),
}));

vi.mock("~/stores/app", () => ({
  app: storeApp,
  currentFeed: () => fixtures.feed,
  selectPost: vi.fn(),
  toggleStar: vi.fn(),
  setPostFilter: vi.fn(),
  setPostSearch,
  togglePostSearch,
  setPostSortOrder,
  markAllRead: vi.fn(),
  loadMorePosts: vi.fn(),
}));

describe("PostList", () => {
  it("shows the title, the summary and the date of each post", () => {
    const { getByText } = render(() => <PostList />);

    expect(getByText("第一篇文章")).toBeTruthy();
    expect(getByText("这是摘要")).toBeTruthy();
    expect(getByText("约 1 分钟 · 250 字")).toBeTruthy();
    expect(() => getByText("约 1 分钟 · 0 字")).toThrow("Unable to find an element with the text");
    expect(getByText(/2025/)).toBeTruthy();
    expect(getByText("A 站")).toBeTruthy();
  });

  it("opens a post when its row is clicked", async () => {
    const store = await import("~/stores/app");
    const { getByText } = render(() => <PostList />);

    fireEvent.click(getByText("第一篇文章"));

    expect(store.selectPost).toHaveBeenCalledWith("guid-1", 1);
  });

  it("stars a post without opening it", async () => {
    const store = await import("~/stores/app");
    const { getAllByRole } = render(() => <PostList />);

    const [star] = getAllByRole("button", { name: "加星标" });
    fireEvent.click(star!);

    expect(store.toggleStar).toHaveBeenCalledWith("guid-1", 1);
    expect(store.selectPost).not.toHaveBeenCalled();
  });

  it("offers to load more posts", async () => {
    const store = await import("~/stores/app");
    const { getByText } = render(() => <PostList />);

    fireEvent.click(getByText(/加载更多/));

    expect(store.loadMorePosts).toHaveBeenCalled();
  });

  it("asks for the opposite order when the sort button is clicked", () => {
    const { getByRole } = render(() => <PostList />);

    fireEvent.click(getByRole("button", { name: "按时间升序排列" }));

    expect(setPostSortOrder).toHaveBeenCalledWith("asc");
  });

  it("toggles the title search field and passes its keyword", () => {
    storeApp.postSearchOpen = true;
    const { getByRole } = render(() => <PostList />);

    const input = getByRole("searchbox", { name: "搜索文章标题" });
    fireEvent.input(input, { target: { value: " 标题 " } });

    expect(setPostSearch).toHaveBeenCalledWith(" 标题 ");

    fireEvent.click(getByRole("button", { name: "关闭标题搜索" }));
    expect(togglePostSearch).toHaveBeenCalledTimes(1);
    storeApp.postSearchOpen = false;
  });
});
