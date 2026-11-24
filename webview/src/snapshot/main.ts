import { Posts, Settings, Snapshot } from "../mygo";
import type { Post, SnapshotGeometry, SnapshotTile } from "../mygo";
import { isMyGo } from "mygo-runtime";

import "./styles.css";

if (!isMyGo()) {
  throw new Error("mygo runtime not found");
}

const mustGet = <T extends HTMLElement>(id: string): T => {
  const element = document.getElementById(id);
  if (!element) throw new Error(`missing element #${id}`);
  return element as T;
};

const elements = {
  stage: mustGet<HTMLElement>("stage"),
  veil: mustGet<HTMLDivElement>("veil"),
  hud: mustGet<HTMLDivElement>("hud"),
  hudText: mustGet<HTMLParagraphElement>("hud-text"),
  error: mustGet<HTMLElement>("error"),
  errorText: mustGet<HTMLParagraphElement>("error-text"),
  errorClose: mustGet<HTMLButtonElement>("error-close"),
};

const maxSnapshotTiles = 32;

const messageOf = (error: unknown): string => {
  if (error instanceof Error && error.message) return error.message;
  return String(error ?? "未知错误");
};

const setBusy = (busy: boolean, text?: string) => {
  elements.veil.hidden = !busy;
  elements.hud.hidden = !busy;
  if (text) elements.hudText.textContent = text;
};

const waitForPaint = () => new Promise<void>((resolve) => requestAnimationFrame(() => resolve()));

const waitForImages = async (card: HTMLElement) => {
  const images = [...card.querySelectorAll("img")];
  await Promise.all(
    images.map(async (image) => {
      image.loading = "eager";
      image.decoding = "sync";
      try {
        await image.decode();
      } catch {
        image.remove();
      }
    }),
  );
};

const captureOffsets = (contentHeight: number, tileHeight: number): number[] => {
  const step = Math.max(1, tileHeight - 1);
  const offsets = [0];
  let offset = 0;
  while (offset + tileHeight < contentHeight) {
    const next = Math.min(offset + step, contentHeight - tileHeight);
    if (next <= offset) break;
    offsets.push(next);
    offset = next;
  }
  if (offsets.length > maxSnapshotTiles) {
    throw new Error("文章太长，无法生成一张快照");
  }
  return offsets;
};

const formatDateTime = (value: string): string => {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "long",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(date);
};

const buildCard = (post: Post, feedTitle: string): HTMLElement => {
  const card = document.createElement("article");
  card.className = "snapshot-card";

  const header = document.createElement("header");
  const meta = document.createElement("p");
  meta.className = "snapshot-meta";
  meta.textContent = [feedTitle.trim(), formatDateTime(post.publishedAt)]
    .filter(Boolean)
    .join(" · ");
  const heading = document.createElement("h1");
  heading.className = "snapshot-title";
  heading.textContent = post.title;
  header.append(meta, heading);
  if (post.author.trim()) {
    const author = document.createElement("p");
    author.className = "snapshot-meta snapshot-author";
    author.textContent = post.author;
    header.append(author);
  }

  // Post.Content is sanitized HTML from the database; Post.Summary is plain
  // text and must never be parsed as markup.
  const body = document.createElement("div");
  body.className = "post-body";
  if (post.content) {
    body.innerHTML = post.content;
  } else {
    body.textContent = post.summary;
  }

  const footer = document.createElement("footer");
  if (post.url.trim()) {
    const source = document.createElement("p");
    source.className = "snapshot-source";
    source.textContent = post.url;
    footer.append(source);
  }
  const credit = document.createElement("p");
  credit.className = "snapshot-credit";
  credit.textContent = "来自 Ressu";
  footer.append(credit);

  card.append(header, body, footer);
  return card;
};

const captureTiles = async (card: HTMLElement): Promise<SnapshotTile[]> => {
  setBusy(true, "正在准备文章快照…");
  await waitForPaint();
  await waitForImages(card);
  await waitForPaint();
  await waitForPaint();

  const viewportWidth = Math.max(1, Math.round(window.innerWidth));
  const cardWidth = Math.round(card.getBoundingClientRect().width);
  const cardLeft = Math.max(0, Math.round((viewportWidth - cardWidth) / 2));
  const contentHeight = elements.stage.scrollHeight;
  const tileHeight = elements.stage.clientHeight;
  if (contentHeight <= 0 || tileHeight <= 0 || cardWidth <= 0) {
    throw new Error("无法准备文章快照");
  }
  const geometry: SnapshotGeometry = {
    viewportWidth,
    viewportHeight: tileHeight,
    cardWidth,
    cardLeft,
    contentHeight,
  };
  const offsets = captureOffsets(contentHeight, tileHeight);
  const tiles: SnapshotTile[] = [];
  let completed = 0;
  for (const offset of offsets) {
    // The HUD is a fixed overlay, so a capture would include it: keep it
    // hidden for the scroll-and-capture beat and show it between tiles to
    // report progress.
    elements.hud.hidden = true;
    elements.stage.scrollTop = offset;
    await waitForPaint();
    await waitForPaint();

    const actualOffset = Math.round(elements.stage.scrollTop);
    if (Math.abs(elements.stage.scrollHeight - contentHeight) > 1) {
      throw new Error("文章在生成快照时发生了变化");
    }
    const png = await Posts.captureSnapshot({ offset: actualOffset, geometry });
    tiles.push({ offset: actualOffset, png, geometry });
    completed += 1;
    setBusy(true, `正在生成文章快照… ${completed}/${offsets.length}`);
  }
  return tiles;
};

const fail = (error: unknown) => {
  setBusy(false);
  elements.errorText.textContent = messageOf(error);
  elements.error.hidden = false;
};

// Preset article font family keywords resolve to the same stacks as the
// Go side; any other value is a custom CSS font-family list used as-is.
const articleFontFamilyStacks: Record<string, string> = {
  serif: 'Georgia, "Times New Roman", "Songti SC", "SimSun", serif',
  sans: 'system-ui, -apple-system, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
  mono: 'ui-monospace, "Cascadia Mono", Consolas, "Sarasa Mono SC", monospace',
};

// The snapshot card follows the reading preferences so the captured
// image reads like the article view.
const applyReadingPrefs = async () => {
  const prefs = await Settings.get();
  const style = document.documentElement.style;
  const family = prefs.articleFontFamily || "";
  if (!family || family === "system") {
    style.removeProperty("--reader-font-family");
  } else if (articleFontFamilyStacks[family]) {
    style.setProperty("--reader-font-family", articleFontFamilyStacks[family]);
  } else {
    style.setProperty("--reader-font-family", family);
  }
  const size = Number(prefs.articleFontSize) || 0;
  if (size) {
    style.setProperty("--reader-font-size", `${size}px`);
  }
  const lineHeight = Number(prefs.articleLineHeight) || 0;
  if (lineHeight) {
    style.setProperty("--reader-line-height", String(lineHeight));
  }
};

const run = async () => {
  try {
    const target = await Snapshot.selected();
    if (!target || !target.guid) {
      throw new Error("没有选中的文章");
    }
    const post = await Posts.get(target.feedId, target.guid);
    if (!post.content && !post.summary) {
      throw new Error("这篇文章没有可以生成快照的内容");
    }
    await applyReadingPrefs();
    const card = buildCard(post, target.feedTitle || "");
    elements.stage.append(card);
    const tiles = await captureTiles(card);
    setBusy(true, "正在合成文章快照…");
    await Posts.stitchSnapshot(post.title || "", tiles);
    // A successful stitch opens the native preview and closes this window.
  } catch (error) {
    fail(error);
  }
};

const closeWindow = () => {
  Snapshot.cancel().catch(() => {});
};

window.addEventListener("keydown", (event) => {
  if (event.key !== "Escape" || event.isComposing) return;
  event.preventDefault();
  closeWindow();
});
elements.errorClose.addEventListener("click", closeWindow);
window.addEventListener("unload", () => {
  elements.errorClose.removeEventListener("click", closeWindow);
});

void run();
