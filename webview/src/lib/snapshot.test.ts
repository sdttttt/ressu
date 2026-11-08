import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import type { Post } from "~/mygo";

const { captureSnapshot } = vi.hoisted(() => ({ captureSnapshot: vi.fn() }));

vi.mock("~/mygo", () => ({
  Posts: { captureSnapshot },
}));

import { generateSnapshot } from "~/lib/snapshot";

const post: Post = {
  feedId: 1,
  guid: "guid-1",
  title: "第一篇文章",
  url: "https://a.example/1",
  author: "作者",
  summary: "",
  wordCount: 3,
  content: '<p>正文</p><img src="https://img.example/1.png">',
  publishedAt: "2026-01-02T00:00:00Z",
  fetchedAt: "2026-01-02T00:00:00Z",
  read: true,
  starred: false,
};

/** The card mounted for capture, kept while the capture is in flight. */
let captured: HTMLElement | undefined;
let imageEvents: string[];

beforeEach(() => {
  captureSnapshot.mockReset();
  captured = undefined;
  imageEvents = [];
  vi.stubGlobal("requestAnimationFrame", (callback: FrameRequestCallback) => {
    setTimeout(() => callback(performance.now()), 0);
    return 0;
  });
  HTMLImageElement.prototype.decode = () => {
    imageEvents.push("decode");
    return Promise.resolve();
  };
  Object.defineProperty(HTMLElement.prototype, "clientHeight", {
    configurable: true,
    get() {
      return this.classList.contains("snapshot-scroll") ? 100 : 0;
    },
  });
  Object.defineProperty(HTMLElement.prototype, "scrollHeight", {
    configurable: true,
    get() {
      return this.classList.contains("snapshot-scroll") ? 250 : 0;
    },
  });
  captureSnapshot.mockImplementation(async () => {
    imageEvents.push("capture");
    return "cG5n";
  });
});

afterEach(() => {
  vi.unstubAllGlobals();
  Reflect.deleteProperty(HTMLElement.prototype, "clientHeight");
  Reflect.deleteProperty(HTMLElement.prototype, "scrollHeight");
});

describe("generateSnapshot", () => {
  it("mounts the card, captures its scroll stops and cleans up", async () => {
    const tiles = await generateSnapshot(post, "A 站");

    expect(tiles).toHaveLength(3);
    expect(tiles.map((tile) => tile.offset)).toEqual([0, 99, 150]);
    expect(tiles[0]!.geometry).toEqual({
      viewportWidth: 1024,
      viewportHeight: 100,
      cardWidth: 760,
      contentHeight: 250,
    });
    expect(captureSnapshot).toHaveBeenCalledTimes(3);
    expect(imageEvents.slice(0, 2)).toEqual(["decode", "capture"]);
    expect(captureSnapshot).toHaveBeenNthCalledWith(1, {
      offset: 0,
      geometry: {
        viewportWidth: 1024,
        viewportHeight: 100,
        cardWidth: 760,
        contentHeight: 250,
      },
    });
    expect(captureSnapshot).toHaveBeenNthCalledWith(3, {
      offset: 150,
      geometry: {
        viewportWidth: 1024,
        viewportHeight: 100,
        cardWidth: 760,
        contentHeight: 250,
      },
    });
    expect(document.querySelector(".snapshot-card")).toBeNull();
    expect(captured).toBeUndefined();
  });

  it("builds a sanitized light card with metadata and source", async () => {
    const promise = generateSnapshot(post, "A 站");
    captured = document.querySelector(".snapshot-card") ?? undefined;

    await promise;
    expect(captured).toBeDefined();
    expect(captured!.querySelector(".snapshot-title")!.textContent).toBe("第一篇文章");
    expect(captured!.querySelector(".snapshot-meta")!.textContent).toContain("A 站");
    expect(captured!.querySelector(".snapshot-author")!.textContent).toBe("作者");
    expect(captured!.querySelector(".post-body")!.textContent).toContain("正文");
    expect(captured!.textContent).toContain("https://a.example/1");
    expect(captured!.textContent).toContain("来自 Ressu");
  });

  it("refuses a post without any body", async () => {
    await expect(generateSnapshot({ ...post, content: "", summary: "" }, "A 站")).rejects.toThrow(
      "这篇文章没有可以生成快照的内容",
    );
    expect(document.querySelector(".snapshot-card")).toBeNull();
  });
});
