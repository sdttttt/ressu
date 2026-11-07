import { formatDateTime } from "~/lib/format";
import { sanitizeHTML } from "~/lib/sanitize";
import { Posts, type Post, type SnapshotTile } from "~/mygo";

/** The fixed CSS width of the capture. It must match the snapshot CSS. */
const CARD_WIDTH = 760;

/** The most visible-page captures one article may use. */
const MAX_TILES = 32;

/**
 * Captures a post through the system WebView and returns cropped tiles.
 * The page mounts a light-themed card in the left part of the viewport,
 * scrolls it, and asks Go to capture the visible page at each stop; Go
 * crops the card and stitches the captures into one PNG.
 */
export async function generateSnapshot(post: Post, feedTitle: string): Promise<SnapshotTile[]> {
  if (!post.content && !post.summary) {
    throw new Error("这篇文章没有可以生成快照的内容");
  }

  const holder = document.createElement("div");
  holder.setAttribute("aria-hidden", "true");
  holder.className = "snapshot-overlay";

  const scroller = document.createElement("div");
  scroller.className = "snapshot-scroll";
  scroller.style.width = `${CARD_WIDTH}px`;
  const card = buildCard(post, feedTitle);
  scroller.append(card);
  holder.append(scroller);
  document.body.append(holder);

  try {
    await waitForPaint();
    await waitForImages(card);
    await waitForPaint();
    await waitForPaint();
    const contentHeight = scroller.scrollHeight;
    const tileHeight = scroller.clientHeight;
    if (contentHeight <= 0 || tileHeight <= 0) {
      throw new Error("无法准备文章快照");
    }

    const viewportWidth = window.innerWidth;
    if (viewportWidth < CARD_WIDTH) {
      throw new Error("窗口太窄，无法生成快照");
    }

    const geometry = {
      viewportWidth,
      viewportHeight: tileHeight,
      cardWidth: CARD_WIDTH,
      contentHeight,
    };
    const offsets = captureOffsets(contentHeight, tileHeight);
    const tiles: SnapshotTile[] = [];
    for (const offset of offsets) {
      scroller.scrollTop = offset;
      await waitForPaint();
      await waitForPaint();

      const actualOffset = Math.round(scroller.scrollTop);
      const currentHeight = scroller.scrollHeight;
      if (Math.abs(currentHeight - contentHeight) > 1) {
        throw new Error("文章在生成快照时发生了变化");
      }
      const png = await Posts.captureSnapshot({ offset: actualOffset, geometry });
      tiles.push({ offset: actualOffset, png, geometry });
    }
    return tiles;
  } finally {
    holder.remove();
  }
}

/** Builds the offscreen card. Text goes in through textContent; only
 * sanitized HTML reaches innerHTML. */
function buildCard(post: Post, feedTitle: string): HTMLElement {
  const card = document.createElement("article");
  card.className = "snapshot-card";

  const meta = document.createElement("p");
  meta.className = "snapshot-meta";
  meta.textContent = [feedTitle.trim(), formatDateTime(post.publishedAt)]
    .filter(Boolean)
    .join(" · ");

  const title = document.createElement("h1");
  title.className = "snapshot-title";
  title.textContent = post.title;

  const header = document.createElement("header");
  header.append(meta, title);
  if (post.author.trim()) {
    const author = document.createElement("p");
    author.className = "snapshot-meta snapshot-author";
    author.textContent = post.author;
    header.append(author);
  }

  const body = document.createElement("div");
  body.className = "post-body";
  // The HTML is cleaned by DOMPurify before it reaches the DOM.
  body.innerHTML = sanitizeHTML(post.content || post.summary, post.url);

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
}

/** Chooses overlapping scroll stops from top to bottom. */
function captureOffsets(contentHeight: number, tileHeight: number): number[] {
  const step = Math.max(1, tileHeight - 1);
  const offsets = [0];
  let offset = 0;
  while (offset + tileHeight < contentHeight) {
    const next = Math.min(offset + step, contentHeight - tileHeight);
    if (next <= offset) break;
    offsets.push(next);
    offset = next;
  }
  if (offsets.length > MAX_TILES) {
    throw new Error("文章太长，无法生成一张快照");
  }
  return offsets;
}

/** Loads card images eagerly; a failed image is removed instead of leaving
 * a broken frame in the shared picture. */
async function waitForImages(card: HTMLElement): Promise<void> {
  const images = [...card.querySelectorAll<HTMLImageElement>("img")];
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
}

/** Waits long enough for layout and paint to settle before a native capture. */
function waitForPaint(): Promise<void> {
  return new Promise((resolve) => requestAnimationFrame(() => resolve()));
}
