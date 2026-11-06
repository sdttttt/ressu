import { render } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";

import PostView from "~/components/post-view";

const feed = { id: 1, title: "A 站" };

// The mock store lives in the factory's module instance, which runs in the
// same ESM graph as the component, so reactive writes propagate.
vi.mock("~/stores/app", async () => {
  const { mockStore, __mockBridge } = await import("~/test/app-mock");
  return {
    app: mockStore,
    currentFeed: () => feed,
    openLink: vi.fn(),
    setOpenPostRead: vi.fn(),
    toggleStar: vi.fn(),
    __mockBridge,
  };
});

const { __mockBridge } = (await import("~/stores/app")) as unknown as {
  __mockBridge: { open: (guid: string) => void };
};

async function openPost(guid: string) {
  __mockBridge.open(guid);
  // 让 effect 链路完成;覆盖层在 300ms 后移除,断言窗口在其之前
  await new Promise((r) => setTimeout(r, 40));
}

/** The data-layer attribute of each rendered article layer. */
function layers(container: HTMLElement): (string | undefined)[] {
  return Array.from(container.querySelectorAll("[data-layer]")).map(
    (el) => (el as HTMLElement).dataset.layer,
  );
}

describe("PostView cover transition", () => {
  it("mounts the first article with the enter animation", async () => {
    const { container } = render(() => <PostView />);
    await openPost("guid-1");

    expect(layers(container)).toEqual(["enter"]);
    const layer = container.querySelector("[data-layer='enter']");
    expect(layer!.className).toContain("slide-in-from-bottom-6");
  });

  it("covers with a top-down sheet when opening a later post", async () => {
    const { container } = render(() => <PostView />);
    await openPost("guid-1");
    await openPost("guid-3"); // 列表里更靠后 → next → 从上方盖下

    expect(layers(container)).toEqual(["base", "next"]);
    expect(container.querySelector("[data-layer='next']")!.className).toContain(
      "slide-in-from-top-full",
    );
    expect(container.textContent).toContain("文章 guid-1");
    expect(container.textContent).toContain("文章 guid-3");
  });

  it("covers with a bottom-up sheet when opening an earlier post", async () => {
    const { container } = render(() => <PostView />);
    await openPost("guid-3");
    await openPost("guid-1"); // 往回看 → prev → 从下方盖上

    expect(layers(container)).toEqual(["base", "prev"]);
    expect(container.querySelector("[data-layer='prev']")!.className).toContain(
      "slide-in-from-bottom-full",
    );
  });

  it("keeps the covered article in the DOM until the slide is done", async () => {
    const { container } = render(() => <PostView />);
    await openPost("guid-1");
    await openPost("guid-2");

    expect(container.textContent).toContain("文章 guid-1");
    expect(container.textContent).toContain("文章 guid-2");
    const cover = container.querySelector("[data-layer='next']");
    expect(cover).toBeTruthy();

    // 300ms 后 base 层被移除;incoming 层必须是同一个 DOM 节点 —— 重建它会
    // 把滑入动画重播一遍,看起来就像连续覆盖了两次
    await new Promise((r) => setTimeout(r, 350));
    expect(layers(container)).toEqual(["next"]);
    expect(container.querySelector("[data-layer='next']")).toBe(cover);
    expect(container.textContent).not.toContain("文章 guid-1");
  });
});
