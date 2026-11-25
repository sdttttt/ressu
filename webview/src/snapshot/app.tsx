import { Show, createEffect, onSettled } from "solid-js";
import { createSignal } from "solid-js";

import { Posts, Settings, Snapshot } from "../mygo";
import type { Post, SnapshotGeometry, SnapshotTile } from "../mygo";
import { captureOffsets } from "./offsets";

const messageOf = (error: unknown): string => {
  if (error instanceof Error && error.message) return error.message;
  return String(error ?? "未知错误");
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

/**
 * Post.Content is sanitized HTML from the database and goes straight into
 * innerHTML; Post.Summary is plain text and must never be parsed as
 * markup.
 */
function CardBody(props: { post: Post }) {
  let element!: HTMLDivElement;
  createEffect(
    () => props.post,
    (post) => {
      if (post.content) {
        element.innerHTML = post.content;
      } else {
        element.textContent = post.summary;
      }
    },
  );
  return <div class="post-content" ref={element} />;
}

/** The fixed-light snapshot card the capture window photographs. */
function SnapshotCard(props: {
  post: Post;
  feedTitle: string;
  ref: (element: HTMLElement) => void;
}) {
  const meta = () =>
    [props.feedTitle.trim(), formatDateTime(props.post.publishedAt)].filter(Boolean).join(" · ");
  return (
    <article class="snapshot-card" ref={props.ref}>
      <header>
        <p class="snapshot-meta">{meta()}</p>
        <h1 class="snapshot-title">{props.post.title}</h1>
        <Show when={props.post.author.trim()}>
          <p class="snapshot-meta snapshot-author">{props.post.author}</p>
        </Show>
      </header>
      <div class="post-body">
        <CardBody post={props.post} />
      </div>
      <footer>
        <Show when={props.post.url.trim()}>
          <p class="snapshot-source">{props.post.url}</p>
        </Show>
        <p class="snapshot-credit">来自 Ressu</p>
      </footer>
    </article>
  );
}

/** The snapshot capture page: renders the card, photographs it tile by
 *  tile, and hands the tiles to the native stitcher. */
export function SnapshotApp() {
  let stageEl!: HTMLElement;
  let cardEl: HTMLElement | undefined;

  const [busy, setBusy] = createSignal(true);
  const [busyText, setBusyText] = createSignal("正在准备文章快照…");
  const [hudVisible, setHudVisible] = createSignal(true);
  const [error, setError] = createSignal<string | null>(null);
  const [target, setTarget] = createSignal<{ post: Post; feedTitle: string } | null>(null);

  const fail = (cause: unknown) => {
    setBusy(false);
    setError(messageOf(cause));
  };

  const captureTiles = async (card: HTMLElement): Promise<SnapshotTile[]> => {
    setBusy(true);
    setBusyText("正在准备文章快照…");
    await waitForPaint();
    await waitForImages(card);
    await waitForPaint();
    await waitForPaint();

    const viewportWidth = Math.max(1, Math.round(window.innerWidth));
    const cardWidth = Math.round(card.getBoundingClientRect().width);
    const cardLeft = Math.max(0, Math.round((viewportWidth - cardWidth) / 2));
    const contentHeight = stageEl.scrollHeight;
    const tileHeight = stageEl.clientHeight;
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
      setHudVisible(false);
      stageEl.scrollTop = offset;
      await waitForPaint();
      await waitForPaint();

      const actualOffset = Math.round(stageEl.scrollTop);
      if (Math.abs(stageEl.scrollHeight - contentHeight) > 1) {
        throw new Error("文章在生成快照时发生了变化");
      }
      const png = await Posts.captureSnapshot({ offset: actualOffset, geometry });
      tiles.push({ offset: actualOffset, png, geometry });
      completed += 1;
      setHudVisible(true);
      setBusyText(`正在生成文章快照… ${completed}/${offsets.length}`);
    }
    return tiles;
  };

  const run = async () => {
    try {
      const selected = await Snapshot.selected();
      if (!selected || !selected.guid) {
        throw new Error("没有选中的文章");
      }
      const loaded = await Posts.get(selected.feedId, selected.guid);
      if (!loaded.content && !loaded.summary) {
        throw new Error("这篇文章没有可以生成快照的内容");
      }
      await applyReadingPrefs();
      setTarget({ post: loaded, feedTitle: selected.feedTitle || "" });
      // Let the card mount before measuring it.
      await waitForPaint();
      if (!cardEl) {
        throw new Error("无法准备文章快照");
      }
      const tiles = await captureTiles(cardEl);
      setBusy(true);
      setBusyText("正在合成文章快照…");
      await Posts.stitchSnapshot(loaded.title || "", tiles);
      // A successful stitch opens the native preview and closes this window.
    } catch (cause) {
      fail(cause);
    }
  };

  const closeWindow = () => {
    Snapshot.cancel().catch(() => {});
  };

  const handleKey = (event: KeyboardEvent) => {
    if (event.key !== "Escape" || event.isComposing) return;
    event.preventDefault();
    closeWindow();
  };

  onSettled(() => {
    window.addEventListener("keydown", handleKey);
    void run();
    return () => {
      window.removeEventListener("keydown", handleKey);
    };
  });

  return (
    <>
      <main id="stage" aria-live="polite" ref={stageEl}>
        <Show when={target()}>
          {(current) => (
            <SnapshotCard
              post={current().post}
              feedTitle={current().feedTitle}
              ref={(element) => {
                cardEl = element;
              }}
            />
          )}
        </Show>
      </main>
      <div id="veil" hidden={!busy()} />
      <div id="hud" class="hud" hidden={!busy() || !hudVisible()}>
        <span class="spinner" aria-hidden="true" />
        <p id="hud-text">{busyText()}</p>
      </div>
      <section id="error" class="error" hidden={error() === null}>
        <p id="error-text">{error()}</p>
        <button id="error-close" type="button" onClick={closeWindow}>
          关闭
        </button>
      </section>
    </>
  );
}
