import { fireEvent, render, screen } from "@solidjs/testing-library";
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
    saveSnapshot: vi.fn(),
    copySnapshot: vi.fn(),
    mockBridge,
  };
});

const { mockBridge } = (await import("~/stores/app")) as unknown as {
  mockBridge: {
    open: (guid: string) => void;
    clear: () => void;
    setSnapshotting: (value: boolean) => void;
  };
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

  it("positions every layer absolutely so the article scroller is constrained", async () => {
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-1");

    // A statically positioned layer grows past the clipped pane and its
    // inner overflow-y-auto never scrolls.
    const layer = container.querySelector("[data-layer='enter']");
    expect(layer!.className).toContain("absolute inset-0");
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

describe("PostView snapshot menu", () => {
  it("offers save and copy actions for the open article", async () => {
    const store = await import("~/stores/app");
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-1");

    const trigger = container.querySelector('button[aria-label="生成文章快照"]');
    expect(trigger).not.toBeNull();
    fireEvent.pointerDown(trigger!);
    const save = await screen.findByText("保存为图片…");
    fireEvent.pointerUp(save);

    expect(store.saveSnapshot).toHaveBeenCalledTimes(1);
    expect(store.copySnapshot).not.toHaveBeenCalled();
  });

  it("copies through the second menu item", async () => {
    const store = await import("~/stores/app");
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-1");

    fireEvent.pointerDown(container.querySelector('button[aria-label="生成文章快照"]')!);
    const copy = await screen.findByText("复制到剪贴板");
    fireEvent.pointerUp(copy);

    expect(store.copySnapshot).toHaveBeenCalledTimes(1);
    expect(store.saveSnapshot).not.toHaveBeenCalled();
  });

  it("disables the trigger while a snapshot is being generated", async () => {
    const { container } = render(() => <PostView />);
    await closePost();
    await openPost("guid-1");

    const trigger = () => container.querySelector('button[aria-label="生成文章快照"]')!;
    mockBridge.setSnapshotting(true);
    expect((trigger() as HTMLButtonElement).disabled).toBe(true);
    mockBridge.setSnapshotting(false);
    expect((trigger() as HTMLButtonElement).disabled).toBe(false);
  });
});
