import {
  app,
  moveSelection,
  openDialog,
  openSelectedPost,
  openSelectedPostInBrowser,
  refreshAll,
  selectNeighbourFeed,
  toggleOpenPostRead,
  toggleStar,
} from "~/stores/app";

/** A key binding: a bare key, or the same key held with the platform
 * modifier (Cmd on macOS, Ctrl elsewhere). */
export type Binding = { key: string; mod?: boolean };

/** A group heading shown in the shortcuts help dialog. */
export type CommandGroup = "应用" | "文章列表" | "阅读" | "订阅源";

export type Command = {
  id: string;
  binding: Binding;
  group: CommandGroup;
  label: string;
  run: () => void;
};

/** Whether the app runs on macOS, where the system menu owns the mod-key
 * shortcuts and the page should not register them again. */
export const isMac = navigator.userAgent.includes("Mac");

const scrollBoxOf = ():
  | (HTMLElement & {
      scrollTop: number;
      clientHeight: number;
      scrollHeight: number;
    })
  | null => {
  // Layers of an article swap keep their nodes around briefly, so the last
  // match is the topmost — the one actually on screen.
  const boxes = document.querySelectorAll<HTMLElement>("[data-post-scroll]");
  const box = boxes[boxes.length - 1];
  return box ?? null;
};

const scrollOrNext = (): void => {
  const box = scrollBoxOf();
  // No article pane yet: the first press starts the pipeline by opening the
  // first post. A pane still mounting has no height: wait.
  if (!box) {
    void moveSelection(1);
    return;
  }
  if (box.clientHeight === 0) return;

  const bottomGap = box.scrollHeight - box.clientHeight - box.scrollTop;
  if (bottomGap > 1) {
    box.scrollBy({ top: box.clientHeight * 0.9, behavior: "smooth" });
    return;
  }
  // The article is read to the bottom: continue with the next one, the
  // NetNewsWire "space bar pipeline".
  void moveSelection(1);
};

const starSelected = (): void => {
  const post = app.posts.find(
    (item) => item.feedId === app.current.postFeedId && item.guid === app.current.guid,
  );
  if (post) void toggleStar(post.guid, post.feedId);
};

/**
 * Every shortcut the page handles, in help-dialog order. The dialog renders
 * straight from this array, so the list can never drift from the bindings.
 */
export const commands: Command[] = [
  {
    id: "add-feed",
    binding: { key: "N", mod: true },
    group: "应用",
    label: "添加订阅源",
    run: () => openDialog("addFeed"),
  },
  {
    id: "refresh-all-mod",
    binding: { key: "R", mod: true },
    group: "应用",
    label: "刷新全部订阅源",
    run: () => void refreshAll(),
  },
  {
    id: "open-settings",
    binding: { key: ",", mod: true },
    group: "应用",
    label: "打开设置",
    run: () => openDialog("settings"),
  },
  {
    id: "refresh-all",
    binding: { key: "R" },
    group: "应用",
    label: "刷新全部订阅源",
    run: () => void refreshAll(),
  },
  {
    id: "show-help",
    binding: { key: "?" },
    group: "应用",
    label: "显示快捷键帮助",
    run: () => openDialog("help"),
  },
  {
    id: "post-next",
    binding: { key: "j" },
    group: "文章列表",
    label: "选中下一篇文章并打开",
    run: () => void moveSelection(1),
  },
  {
    id: "post-prev",
    binding: { key: "k" },
    group: "文章列表",
    label: "选中上一篇文章并打开",
    run: () => void moveSelection(-1),
  },
  {
    id: "post-next-unread",
    binding: { key: "n" },
    group: "文章列表",
    label: "跳到下一篇未读",
    run: () => void moveSelection(1, true),
  },
  {
    id: "post-prev-unread",
    binding: { key: "p" },
    group: "文章列表",
    label: "跳到上一篇未读",
    run: () => void moveSelection(-1, true),
  },
  {
    id: "post-open",
    binding: { key: "Enter" },
    group: "文章列表",
    label: "打开选中的文章",
    run: () => void openSelectedPost(),
  },
  {
    id: "post-star",
    binding: { key: "s" },
    group: "文章列表",
    label: "切换星标",
    run: starSelected,
  },
  {
    id: "reader-scroll",
    binding: { key: "Space" },
    group: "阅读",
    label: "向下滚动正文，到底时读下一篇",
    run: scrollOrNext,
  },
  {
    id: "reader-toggle-read",
    binding: { key: "m" },
    group: "阅读",
    label: "切换已读 / 未读",
    run: toggleOpenPostRead,
  },
  {
    id: "reader-open-original",
    binding: { key: "v" },
    group: "阅读",
    label: "在浏览器中打开原文",
    run: openSelectedPostInBrowser,
  },
  {
    id: "feed-prev",
    binding: { key: "h" },
    group: "订阅源",
    label: "上一个订阅源",
    run: () => void selectNeighbourFeed(-1),
  },
  {
    id: "feed-next",
    binding: { key: "l" },
    group: "订阅源",
    label: "下一个订阅源",
    run: () => void selectNeighbourFeed(1),
  },
];

/** Renders an event's key the way the bindings spell it. */
function eventKey(event: KeyboardEvent): string {
  if (event.key === " ") return "Space";
  if (event.key === "Escape") return "Esc";
  return event.key;
}

/** Human-readable binding text for the help dialog. */
export function formatBinding(binding: Binding, onMac = isMac): string {
  const key = binding.key === "Space" ? "空格" : binding.key;
  if (binding.mod) return onMac ? `⌘${key}` : `Ctrl+${key}`;
  return key;
}

function anyDialogOpen(): boolean {
  return Object.values(app.dialog).some((open) => open);
}

/**
 * Registers the global keydown listener. Pass `skipMod` on macOS, where the
 * system menu already handles the mod bindings. Returns the unsubscribe
 * function.
 */
export function installShortcuts(options: { skipMod?: boolean } = {}): () => void {
  const onKeyDown = (event: KeyboardEvent) => {
    if (event.isComposing) return;

    const target = event.target;
    const inField =
      target instanceof HTMLInputElement ||
      target instanceof HTMLTextAreaElement ||
      target instanceof HTMLSelectElement ||
      (target instanceof HTMLElement && target.isContentEditable);
    const modHeld = event.ctrlKey || event.metaKey;

    for (const command of commands) {
      const mod = command.binding.mod === true;
      if (mod && options.skipMod) continue;
      if (mod !== modHeld) continue;
      // Bare bindings want a bare key: Shift shows up in the key itself
      // (R, ?), the other modifiers must be off.
      if (!mod && event.altKey) continue;
      if (inField && !mod) continue;
      // A dialog owns the keyboard: Kobalte closes it with Escape, and the
      // bare shortcuts wait until it is gone.
      if (anyDialogOpen() && !mod) continue;
      if (eventKey(event) !== command.binding.key) continue;

      event.preventDefault();
      command.run();
      return;
    }
  };

  window.addEventListener("keydown", onKeyDown);
  return () => window.removeEventListener("keydown", onKeyDown);
}
