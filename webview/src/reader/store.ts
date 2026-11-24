import { createSignal } from "solid-js";

import { Posts, Settings, Status, events } from "../mygo";
import type {
  Appearance,
  Enclosure,
  Event as AppEvent,
  Post,
  PostSelection,
  Prefs,
} from "../mygo";
import { externalHref } from "./links";

/** How an article layer behaves while the stack swaps. */
export type LayerKind = "base" | "enter" | "cover-down" | "exit-up";
type AnimatedKind = Exclude<LayerKind, "base">;

/** The loaded body of an article layer. */
export interface LayerPost {
  /** Sanitized HTML, or null when only the plain-text summary exists. */
  html: string | null;
  summary: string;
  enclosures: Enclosure[] | null;
}

/**
 * One open article. Solid components render it and attach its DOM nodes;
 * the store drives kind changes and content replacement. Fields that never
 * drive rendering stay plain.
 */
export class Layer {
  readonly selection: PostSelection;
  swap: number;
  retired = false;
  onFinished: (() => void) | null = null;
  element: HTMLDivElement | null = null;
  animation: Animation | null = null;

  readonly kind: () => LayerKind;
  readonly setKind: (kind: LayerKind) => void;
  readonly post: () => LayerPost | null;
  readonly setPost: (post: LayerPost | null) => void;

  constructor(selection: PostSelection, kind: LayerKind, swap: number) {
    this.selection = selection;
    this.swap = swap;
    [this.kind, this.setKind] = createSignal<LayerKind>(kind);
    [this.post, this.setPost] = createSignal<LayerPost | null>(null);
  }
}

interface ReaderState {
  revision: number;
  request: number;
}

const state: ReaderState = { revision: 0, request: 0 };

// Matches store.DefaultPrefs; the real preferences arrive from Settings.get
// at startup and with every prefs.changed event.
const defaultPrefs: Prefs = {
  theme: "system",
  fontSize: "medium",
  refreshOnStart: true,
  markReadOnOpen: true,
  remoteResources: "webview",
  openLinks: "browser",
  postSort: "published",
  stripTrackers: false,
  postsPerFeed: 500,
  imageCacheLimit: 256,
  paneSizes: [18, 30],
  sidebarCollapsed: false,
  reduceMotion: false,
  uiFontFamily: "system",
  articleFontFamily: "system",
  articleFontSize: 0,
  articleLineHeight: 1.78,
};

export const [prefs, setPrefs] = createSignal<Prefs>(defaultPrefs);
export const [post, setPost] = createSignal<Post | null>(null);
export const [layers, setLayers] = createSignal<Layer[]>([]);

// Article swap choreography: 300ms slides, leftover layers dropped after
// twice the slide plus one frame of slack for engines that never report
// the slide as finished.
const slideMs = 300;
const swapCleanupMs = 700;

let swapCounter = 0;

const messageOf = (error: unknown): string => {
  if (error instanceof Error && error.message) return error.message;
  return String(error ?? "未知错误");
};

const notifyError = (message: string, error: unknown) => {
  Status.notify("error", `${message}：${messageOf(error)}`).catch(() => {});
};

const isDark = (theme: string): boolean => {
  if (theme === "dark") return true;
  if (theme === "light" || theme === "sepia") return false;
  return window.matchMedia("(prefers-color-scheme: dark)").matches;
};

// Effective reduce-motion: the app's own preference, or the desktop's
// accessibility setting, which the system webview reports as
// prefers-reduced-motion. Read fresh on each use so a system change
// applies without a page reload.
export const reduceMotion = (): boolean =>
  prefs().reduceMotion ||
  window.matchMedia("(prefers-reduced-motion: reduce)").matches;

const fontScale = (fontSize: string): string => {
  if (fontSize === "large") return "1.3";
  if (fontSize === "medium") return "1.15";
  return "1";
};

// Preset article font family keywords resolve to the same stacks as the
// Go side; any other value is a custom CSS font-family list used as-is.
const articleFontFamilyStacks: Record<string, string> = {
  serif: 'Georgia, "Times New Roman", "Songti SC", "SimSun", serif',
  sans: 'system-ui, -apple-system, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
  mono: 'ui-monospace, "Cascadia Mono", Consolas, "Sarasa Mono SC", monospace',
};

export const applyAppearanceFromPrefs = () => {
  const current = prefs();
  const root = document.documentElement;
  root.classList.toggle("dark", isDark(current.theme));
  root.classList.toggle("theme-sepia", current.theme === "sepia");
  const style = root.style;
  style.setProperty("--app-text-scale", fontScale(current.fontSize));

  const family = current.articleFontFamily || "";
  if (!family || family === "system") {
    style.removeProperty("--reader-font-family");
  } else if (articleFontFamilyStacks[family]) {
    style.setProperty("--reader-font-family", articleFontFamilyStacks[family]);
  } else {
    style.setProperty("--reader-font-family", family);
  }

  const size = Number(current.articleFontSize) || 0;
  if (size) {
    style.setProperty("--reader-font-size", `${size}px`);
  } else {
    style.removeProperty("--reader-font-size");
  }

  const lineHeight = Number(current.articleLineHeight) || 0;
  if (lineHeight) {
    style.setProperty("--reader-line-height", String(lineHeight));
  } else {
    style.removeProperty("--reader-line-height");
  }

  // Font and line-height changes resize the content under the same
  // scroll offset.
  updateProgress();
};

const applyAppearance = (appearance: Appearance | null) => {
  if (!appearance) return;
  setPrefs((current) => ({ ...current, theme: appearance.theme, fontSize: appearance.fontSize }));
  applyAppearanceFromPrefs();
};

const applyPrefs = (next: Prefs | null) => {
  if (!next) return;
  setPrefs(next);
  applyAppearanceFromPrefs();
};

export const topLayer = (): Layer | null => {
  const stack = layers();
  return stack[stack.length - 1] || null;
};

// The progress element registers itself here; the bar is metric-driven, so
// it is updated imperatively rather than through a signal.
let progressEl: HTMLDivElement | null = null;

export const registerProgress = (element: HTMLDivElement | null) => {
  progressEl = element;
};

/**
 * Reading progress: the bar tracks the top layer's scroll. Scroll events
 * do not bubble, so one capture-phase listener on #article covers every
 * layer. An article that fits the pane has no progress to show; the bar
 * hides rather than sitting full. Visible updates ease toward the new
 * ratio (the CSS transition); a bar reappearing after being hidden jumps
 * to its position, and the reduce-motion preference pins every update.
 */
export const updateProgress = () => {
  const bar = progressEl;
  if (!bar) return;
  const layer = topLayer();
  const box = layer && layer.element;
  const total = box ? box.scrollHeight - box.clientHeight : 0;
  if (total <= 1) {
    bar.hidden = true;
    return;
  }
  const ratio = Math.min(Math.max(box!.scrollTop / total, 0), 1);
  const instant = bar.hidden || reduceMotion();
  bar.hidden = false;
  if (instant) {
    bar.style.transition = "none";
  }
  bar.style.transform = `scaleX(${ratio})`;
  if (instant) {
    requestAnimationFrame(() => {
      bar.style.transition = "";
    });
  }
};

/** Stop any in-flight slide and clear the pinned first frame. */
const stopLayerAnimation = (layer: Layer) => {
  if (layer.animation) {
    layer.animation.cancel();
    layer.animation = null;
  }
  if (layer.element) {
    layer.element.style.opacity = "";
    layer.element.style.transform = "";
  }
};

/** Entering rises 24px while fading in; cover-down sheets down over the
 *  old article from above; exit-up slides up and away, fully opaque,
 *  revealing the incoming one beneath. */
const kindFrames: Record<AnimatedKind, Keyframe[]> = {
  enter: [
    { opacity: "0", transform: "translateY(24px)" },
    { opacity: "1", transform: "translateY(0)" },
  ],
  "cover-down": [
    { opacity: "1", transform: "translateY(-100%)" },
    { opacity: "1", transform: "translateY(0)" },
  ],
  "exit-up": [
    { opacity: "1", transform: "translateY(0)" },
    { opacity: "1", transform: "translateY(-100%)" },
  ],
};

/** Static first-frame pins for kinds whose animation starts on a later
 *  frame, so the pre-animation frame cannot flash the end state. */
const kindPins: Record<AnimatedKind, Keyframe> = {
  enter: { opacity: "0", transform: "translateY(24px)" },
  "cover-down": { opacity: "1", transform: "translateY(-100%)" },
  "exit-up": { opacity: "1", transform: "translateY(0)" },
};

/**
 * Play a layer's slide for its current kind, driven by the layer
 * component's effect. The animation starts only once the element is
 * attached: an `animate` call made while the element is still detached
 * leaves the animation permanently pending (no start time, `finished`
 * never resolves) in both WebKit and Chromium, which used to leave the
 * pane black. Deferring to the next frame fixes that; the first keyframe
 * is pinned inline until then. An inline `animation: none` keeps CSS
 * animations from compositing over the slide — WebKit lets a concurrent
 * CSS animation win, which would cancel it. Only the Web Animations API
 * drives these slides; there is no CSS animation on layers.
 */
export const playKind = (layer: Layer, kind: LayerKind) => {
  const element = layer.element;
  if (!element) return;
  const onFinished = layer.onFinished;
  stopLayerAnimation(layer);
  if (kind === "base") return;
  if (reduceMotion()) {
    // Land on the end state at once; the swap's cleanup still runs
    // through onFinished, off the stack like an animation's.
    if (onFinished) queueMicrotask(onFinished);
    return;
  }
  element.style.animation = "none";
  element.style.opacity = kindPins[kind].opacity as string;
  element.style.transform = kindPins[kind].transform as string;
  requestAnimationFrame(() => {
    if (!element.isConnected) return;
    element.style.opacity = "";
    element.style.transform = "";
    const animation = element.animate(kindFrames[kind], {
      duration: slideMs,
      easing: "ease-out",
    });
    layer.animation = animation;
    // A rejected `finished` means the layer was replaced mid-flight; the
    // newer swap rebuilds the stack anyway. Consume the rejection even
    // without a callback so a cancelled slide never surfaces as an
    // unhandled rejection.
    if (onFinished) {
      animation.finished.then(onFinished, () => {});
    } else {
      animation.finished.catch(() => {});
    }
  });
};

const setLayerKind = (layer: Layer, kind: LayerKind, onFinished?: () => void) => {
  layer.onFinished = onFinished ?? null;
  layer.setKind(kind);
};

const removeLayer = (layer: Layer) => {
  setLayers((stack) => stack.filter((item) => item !== layer));
};

/** Drop a leftover layer of one swap once its slide has finished. */
const retireLayer = (layer: Layer, swap: number) => {
  if (!layer.retired || layer.swap !== swap) return;
  stopLayerAnimation(layer);
  removeLayer(layer);
};

/** Backstop for engines that never report the slide as finished. */
const scheduleRetire = (layer: Layer, swap: number) => {
  setTimeout(() => retireLayer(layer, swap), swapCleanupMs);
};

export const clearPost = () => {
  setPost(null);
  state.request += 1;
  for (const layer of layers()) {
    stopLayerAnimation(layer);
  }
  setLayers([]);
  updateProgress();
};

export const openLink = async (href: string) => {
  const url = externalHref(href);
  if (!url) return;
  try {
    await Settings.openLink(url);
  } catch (error) {
    notifyError("打开链接失败", error);
  }
};

const samePost = (first: PostSelection, second: PostSelection): boolean =>
  first.feedId === second.feedId && first.guid === second.guid;

/**
 * Swap the open article. The direction tells how the new selection sits
 * relative to the previous one in the native list: 1 later (the current
 * article slides up and away, revealing the incoming one beneath), -1
 * earlier (the new article sheets down over the current one), 0 unknown
 * (the first article fades in; a covered article with a missing
 * direction falls back to being covered). The entering layer mounts
 * immediately and slides while its body still loads, so quick keyboard
 * runs keep animating instead of waiting on `Posts.get`.
 */
const openPost = (selection: PostSelection, direction: number) => {
  const requestId = ++state.request;
  const covered = topLayer();

  if (covered && samePost(covered.selection, selection)) {
    return;
  }

  const swap = ++swapCounter;
  if (covered) {
    covered.retired = true;
    covered.swap = swap;
    if (direction > 0) {
      setLayerKind(covered, "exit-up", () => retireLayer(covered, swap));
    } else {
      setLayerKind(covered, "base");
    }
    scheduleRetire(covered, swap);
  }

  const entering = new Layer(selection, direction > 0 ? "enter" : "cover-down", swap);
  setLayers((stack) => [...stack, entering]);
  updateProgress();
  void loadIntoLayer(entering, requestId);
};

const loadIntoLayer = async (layer: Layer, requestId: number) => {
  try {
    const loaded = await Posts.get(layer.selection.feedId, layer.selection.guid);
    if (requestId !== state.request) return;
    if (!layers().includes(layer)) return;
    // Post.Content is sanitized HTML; Post.Summary is plain text that must
    // never be parsed as markup.
    layer.setPost({
      html: loaded.content || null,
      summary: loaded.summary,
      enclosures: loaded.enclosures,
    });
    setPost(loaded);
    updateProgress();
  } catch (error) {
    if (requestId !== state.request) return;
    restoreCovered(layer);
    notifyError("无法打开文章", error);
  }
};

/** Loading failed: drop the entering layer and bring the covered one back. */
const restoreCovered = (layer: Layer) => {
  stopLayerAnimation(layer);
  removeLayer(layer);
  const previous = topLayer();
  if (previous && previous.retired) {
    previous.retired = false;
    previous.swap = 0;
    setLayerKind(previous, "base");
  }
  updateProgress();
};

export const scrollOrNext = () => {
  const layer = topLayer();
  if (!layer || !post()) {
    Posts.selectNeighbour(1, false).catch(() => {});
    return;
  }
  const box = layer.element;
  if (!box || box.clientHeight === 0) return;
  const bottomGap = box.scrollHeight - box.clientHeight - box.scrollTop;
  if (bottomGap > 1) {
    box.scrollBy({
      top: box.clientHeight * 0.9,
      behavior: reduceMotion() ? "auto" : "smooth",
    });
    return;
  }
  Posts.selectNeighbour(1, false).catch(() => {});
};

export const toggleRead = () => {
  const current = post();
  if (!current) return;
  Posts.toggleRead(current.feedId, current.guid).catch((error) => {
    notifyError("更新阅读状态失败", error);
  });
};

export const selectNeighbourFeed = (delta: number) => {
  Status.selectNeighbourFeed(delta).catch((error) => {
    notifyError("无法切换订阅源", error);
  });
};

export const showHelp = () => {
  Status.showHelp().catch((error) => {
    notifyError("无法打开快捷键帮助", error);
  });
};

export const notifyOpenError = (message: string, error: unknown) => {
  notifyError(message, error);
};

const handleEvent = (event: AppEvent) => {
  if (event.revision <= state.revision) return;
  state.revision = event.revision;

  switch (event.eid) {
    case "state.snapshot":
      if (event.appearance) applyAppearance(event.appearance);
      break;
    case "appearance.changed":
      applyAppearance(event.appearance);
      break;
    case "prefs.changed":
      applyPrefs(event.prefs);
      break;
    case "selection.changed":
      clearPost();
      break;
    case "post.selected":
      if (event.postSelection) {
        openPost(event.postSelection, event.direction || 0);
      }
      break;
    case "post.fulltext": {
      // FullText content is cleaned on the Go side, like Post.Content.
      const fullText = event.fullText;
      const layer = topLayer();
      if (!fullText || !layer || !samePost(layer.selection, fullText)) break;
      const current = layer.post();
      if (current) {
        layer.setPost({ ...current, html: fullText.content });
      }
      const open = post();
      if (open && samePost(open, fullText)) {
        setPost({ ...open, content: fullText.content });
      }
      updateProgress();
      break;
    }
  }
};

/** Wire the store to the runtime: events, preferences, page-ready. */
export const startStore = async () => {
  applyAppearanceFromPrefs();
  const unsubscribe = events.appEvent.on(handleEvent);
  try {
    applyPrefs(await Settings.get());
    await Status.pageReady();
  } catch (error) {
    notifyError("无法读取本地数据", error);
  }
  return unsubscribe;
};
