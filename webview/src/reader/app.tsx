import { For, onSettled } from "solid-js";

import { Posts } from "../mygo";
import { ArticleLayerView } from "./article-layer";
import { openLink } from "./store";
import {
  applyAppearanceFromPrefs,
  layers,
  notifyOpenError,
  post,
  prefs,
  registerProgress,
  scrollOrNext,
  selectNeighbourFeed,
  showHelp,
  startStore,
  toggleRead,
  updateProgress,
} from "./store";

const handleKey = (event: KeyboardEvent) => {
  if (event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return;

  switch (event.key) {
    case " ":
      event.preventDefault();
      scrollOrNext();
      break;
    case "m":
      toggleRead();
      break;
    case "v": {
      const current = post();
      if (current?.url) void openLink(current.url);
      break;
    }
    case "h":
      event.preventDefault();
      void selectNeighbourFeed(-1);
      break;
    case "l":
      event.preventDefault();
      void selectNeighbourFeed(1);
      break;
    case "?":
      event.preventDefault();
      void showHelp();
      break;
  }
};

// Clicks inside article bodies are delegated: images open in the native
// viewer, links pass the external-address filter before reaching the Go
// binding.
const handleBodyClick = (event: MouseEvent) => {
  const target = event.target instanceof Element ? event.target : null;
  if (!target) return;

  const image = target.closest("img");
  if (image) {
    event.preventDefault();
    const source = image.currentSrc || image.getAttribute("src") || "";
    const alt = image.getAttribute("alt") || "";
    Posts.openImage(source, alt).catch((error) => {
      notifyOpenError("无法打开图片", error);
    });
    return;
  }

  const anchor = target.closest("a");
  if (!anchor) return;
  const href = anchor.getAttribute("href");
  if (!href) return;
  event.preventDefault();
  void openLink(href);
};

/** The reader page: empty state, the article stack, and the progress bar. */
export function ReaderApp() {
  let stackEl!: HTMLDivElement;

  onSettled(() => {
    let unsubscribe: (() => void) | undefined;
    void startStore().then((stop) => {
      unsubscribe = stop;
    });

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const onSystemThemeChange = () => {
      if (prefs().theme === "system") applyAppearanceFromPrefs();
    };
    mediaQuery.addEventListener("change", onSystemThemeChange);
    window.addEventListener("keydown", handleKey);
    stackEl.addEventListener("scroll", updateProgress, true);
    window.addEventListener("resize", updateProgress);

    return () => {
      unsubscribe?.();
      mediaQuery.removeEventListener("change", onSystemThemeChange);
      window.removeEventListener("keydown", handleKey);
      window.removeEventListener("resize", updateProgress);
    };
  });

  const hasLayers = () => layers().length > 0;

  return (
    <>
      <section id="empty" class="reader-state" hidden={hasLayers()}>
        <svg aria-hidden="true" viewBox="0 0 24 24" fill="none" stroke="currentColor">
          <path d="M4 11a9 9 0 0 1 9 9" />
          <path d="M4 4a16 16 0 0 1 16 16" />
          <circle cx="5" cy="19" r="1" />
        </svg>
        <h2>选择一篇文章</h2>
        <p>左侧列表里的文章会显示在这里。</p>
      </section>
      <div
        id="article"
        hidden={!hasLayers()}
        ref={stackEl}
        onClick={handleBodyClick}
      >
        <For each={layers()}>{(layer) => <ArticleLayerView layer={layer} />}</For>
      </div>
      <div id="progress" hidden ref={(element) => registerProgress(element)} />
    </>
  );
}
