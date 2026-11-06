import { createStore } from "solid-js/store";

// The mock store lives inside the vi.mock factory's module instance (see
// components/post-view.test.tsx) so the component reads the same reactive
// graph it was rendered with; this module only exposes the handles.
export type MockPost = {
  feedId: number;
  guid: string;
  title: string;
  url: string;
  author: string;
  summary: string;
  publishedAt: string;
  read: boolean;
  starred: boolean;
  content: string;
};

const [store, setStore] = createStore({
  post: null as MockPost | null,
  loadingPost: false,
  posts: [] as MockPost[],
});

const posts = [make("guid-1"), make("guid-2"), make("guid-3")];

function make(guid: string): MockPost {
  return {
    feedId: 1,
    guid,
    title: `文章 ${guid}`,
    url: `https://a.example/${guid}`,
    author: "",
    summary: "摘要",
    publishedAt: "2026-01-01T00:00:00Z",
    read: false,
    starred: false,
    content: "<p>正文</p>",
  };
}

export const mockStore = store;

export const __mockBridge = {
  open(guid: string) {
    // Fresh deep copies mirror the mygo bridge, where every call
    // deserializes new objects. Sharing raw objects between `post` and
    // `posts` would make Solid store's key-wise merge write through and
    // corrupt the list (a test-only artifact the real bridge never has).
    setStore(
      "posts",
      posts.map((p) => ({ ...p })),
    );
    setStore("post", { ...posts.find((p) => p.guid === guid)! });
  },
};
