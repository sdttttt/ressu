import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { formatBinding, installShortcuts } from "~/lib/commands";

vi.mock("~/stores/app", () => {
  const state = {
    feeds: [
      { id: 1, title: "A" },
      { id: 2, title: "B" },
    ],
    posts: [{ feedId: 1, guid: "guid-1" }],
    current: { feedId: 1, postFeedId: 1, guid: null },
    post: null,
    dialog: { addFeed: false, help: false },
  };
  return {
    app: state,
    moveSelection: vi.fn(),
    openDialog: vi.fn(),
    openSettings: vi.fn(),
    openSelectedPost: vi.fn(),
    openSelectedPostInBrowser: vi.fn(),
    refreshAll: vi.fn(),
    selectNeighbourFeed: vi.fn(),
    toggleOpenPostRead: vi.fn(),
    toggleStar: vi.fn(),
  };
});

beforeEach(() => {
  vi.clearAllMocks();
});

const store = async () => await import("~/stores/app");

const press = (init: KeyboardInit) => {
  const target = init.target ?? window;
  const event = new KeyboardEvent("keydown", {
    key: init.key,
    ctrlKey: init.ctrl ?? false,
    metaKey: init.meta ?? false,
    shiftKey: init.shift ?? false,
    altKey: init.alt ?? false,
    bubbles: true,
  });
  Object.defineProperty(event, "isComposing", { value: init.composing ?? false });
  target.dispatchEvent(event);
};

type KeyboardInit = {
  key: string;
  target?: EventTarget;
  ctrl?: boolean;
  meta?: boolean;
  shift?: boolean;
  alt?: boolean;
  composing?: boolean;
};

beforeEach(() => {
  document.body.innerHTML = "";
});

afterEach(() => {
  document.body.innerHTML = "";
});

describe("installShortcuts", () => {
  it("runs the command a bare key is bound to", async () => {
    const app = await store();
    const stop = installShortcuts();

    press({ key: "j" });

    expect(app.moveSelection).toHaveBeenCalledWith(1);
    stop();
  });

  it("runs a mod binding only with the modifier held", async () => {
    const app = await store();
    const stop = installShortcuts();

    press({ key: "," });
    expect(app.openSettings).not.toHaveBeenCalled();

    press({ key: ",", ctrl: true });
    expect(app.openSettings).toHaveBeenCalled();
    stop();
  });

  it("marks the event handled for bare keys", () => {
    const stop = installShortcuts();
    const event = new KeyboardEvent("keydown", { key: "s", cancelable: true });

    window.dispatchEvent(event);

    expect(event.defaultPrevented).toBe(true);
    stop();
  });

  it("ignores bare keys while an input has the focus", async () => {
    const app = await store();
    const input = document.createElement("input");
    document.body.append(input);
    input.focus();
    const stop = installShortcuts();

    press({ key: "j", target: input });
    press({ key: "s", target: input });

    expect(app.moveSelection).not.toHaveBeenCalled();
    stop();
  });

  it("keeps mod bindings working inside an input", async () => {
    const app = await store();
    const input = document.createElement("input");
    document.body.append(input);
    input.focus();
    const stop = installShortcuts();

    press({ key: "N", target: input, ctrl: true });

    expect(app.openDialog).toHaveBeenCalledWith("addFeed");
    stop();
  });

  it("ignores bare keys while a dialog is open", async () => {
    const app = await store();
    const stop = installShortcuts();
    const { app: state } = await store();
    state.dialog.help = true;

    press({ key: "j" });
    press({ key: "?" });

    expect(app.moveSelection).not.toHaveBeenCalled();
    stop();

    state.dialog.help = false;
  });

  it("ignores keys during IME composition", async () => {
    const app = await store();
    const stop = installShortcuts();

    press({ key: "j", composing: true });

    expect(app.moveSelection).not.toHaveBeenCalled();
    stop();
  });

  it("leaves mod bindings to the system menu when asked to", async () => {
    const app = await store();
    const stop = installShortcuts({ skipMod: true });

    press({ key: "R", ctrl: true });
    press({ key: "R" });

    expect(app.refreshAll).toHaveBeenCalledTimes(1);
    stop();
  });

  it("does not treat a plain letter with Alt held as a bare key", async () => {
    const app = await store();
    const stop = installShortcuts();

    press({ key: "j", alt: true });

    expect(app.moveSelection).not.toHaveBeenCalled();
    stop();
  });

  it("stars the selected post with s", async () => {
    const app = await store();
    const stop = installShortcuts();
    app.app.current.guid = "guid-1";

    press({ key: "s" });

    expect(app.toggleStar).toHaveBeenCalledWith("guid-1", 1);
    stop();
    app.app.current.guid = null;
  });

  it("shows the help dialog for ?", async () => {
    const app = await store();
    const stop = installShortcuts();

    press({ key: "?", shift: true });

    expect(app.openDialog).toHaveBeenCalledWith("help");
    stop();
  });
});

describe("formatBinding", () => {
  it("renders the platform modifier per platform", () => {
    expect(formatBinding({ key: "N", mod: true }, false)).toBe("Ctrl+N");
    expect(formatBinding({ key: "N", mod: true }, true)).toBe("⌘N");
    expect(formatBinding({ key: "j" })).toBe("j");
  });

  it("spells the space key in Chinese", () => {
    expect(formatBinding({ key: "Space" })).toBe("空格");
  });
});
