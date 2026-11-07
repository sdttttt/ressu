import { render } from "@solidjs/testing-library";
import { describe, expect, it, vi } from "vitest";

import PostView from "~/components/post-view";

const feed = { id: 1, title: "A 站" };

// The mock store lives in the factory's module instance, which runs in the
// same ESM graph as the component, so reactive writes propagate.
vi.mock("~/stores/app", async () => {
  const { mockStore, mockBridge } = await import("~/test/app-mock");
  return {
    app: mockStore,
    currentFeed: () => feed,
    openLink: vi.fn(),
    setOpenPostRead: vi.fn(),
    toggleStar: vi.fn(),
    mockBridge,
  };
});

const { mockBridge } = (await import("~/stores/app")) as unknown as {
  mockBridge: { open: (guid: string) => void; clear: () => void };
};

async function openPost(guid: string) {
  mockBridge.open(guid);
  // 让 effect 链路完成;旧层在动画结束后移除,断言窗口在其之前
  await new Promise((r) => setTimeout(r, 40));
}

/** Resets the open post so each case starts from an empty pane. */
async function closePost() {
  mockBridge.clear();
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
    await closePost();
    await openPost("guid-1");

    expect(layers(container)).toEqual(["enter"]);
    const layer = container.querySelector("[data-layer='enter']");
    expect(layer!.className).toContain("slide-in-from-bottom-6");
  });

  it("slides the current article up when opening a later post", async () => {
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-1");
    const before = container.querySelector("[data-layer='enter']");

    await openPost("guid-3"); // 列表里更靠后 → 当前文章向上滑走,露出下一篇

    expect(layers(container)).toEqual(["exit-up", "enter"]);
    expect(container.textContent).toContain("文章 guid-1");
    expect(container.textContent).toContain("文章 guid-3");
    // 旧层节点被复用:同一个 DOM 节点只是换了个角色
    expect(container.querySelector("[data-layer='exit-up']")).toBe(before);
  });

  it("sheets the previous article down from the top when going back", async () => {
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-3");
    await openPost("guid-1"); // 往回看 → 上一篇从上方盖下来

    expect(layers(container)).toEqual(["base", "cover-down"]);
    expect(container.querySelector("[data-layer='cover-down']")!.className).toContain(
      "slide-in-from-top-full",
    );
  });

  it("keeps the outgoing article in the DOM until the slide is done", async () => {
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-1");
    const before = container.querySelector("[data-layer='enter']");
    await openPost("guid-2");

    expect(container.textContent).toContain("文章 guid-1");
    expect(container.textContent).toContain("文章 guid-2");
    const outgoing = container.querySelector("[data-layer='exit-up']");
    expect(outgoing).toBe(before);

    // 旧层在滑出动画结束时移除(约 300ms);jsdom 没有 WAAPI,由兜底
    // 定时器(COVER_MS*2+100)移除。incoming 层必须是同一个 DOM 节点 ——
    // 重建它会把入场动画重播一遍
    await new Promise((r) => setTimeout(r, 750));
    expect(layers(container)).toEqual(["enter"]);
    expect(container.querySelector("[data-layer='enter']")).not.toBe(before);
    expect(container.textContent).not.toContain("文章 guid-1");
  });
});
