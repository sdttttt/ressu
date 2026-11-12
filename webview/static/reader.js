(() => {
  "use strict";

  const runtime = globalThis.mygo;
  if (!runtime) {
    throw new Error("mygo runtime not found");
  }

  const state = {
    revision: 0,
    post: null,
    request: 0,
    feeds: [],
    prefs: {
      theme: "system",
      fontSize: "medium",
      markReadOnOpen: true,
    },
  };

  const snapshot = {
    busy: false,
    request: 0,
  };

  const elements = {
    article: document.getElementById("article"),
    empty: document.getElementById("empty"),
    snapshot: document.getElementById("snapshot-backdrop"),
    snapshotProgress: document.getElementById("snapshot-progress"),
    snapshotProgressText: document.getElementById("snapshot-progress-text"),
  };

  const maxSnapshotWidth = 760;
  const minSnapshotWidth = 320;
  const maxSnapshotTiles = 32;

  // Article swap choreography: 300ms slides, leftover layers dropped after
  // twice the slide plus one frame of slack for engines that never report
  // the slide as finished.
  const slideMs = 300;
  const swapCleanupMs = 700;

  const messageOf = (error) => {
    if (error instanceof Error && error.message) return error.message;
    return String(error ?? "未知错误");
  };

  const notify = (level, message) => {
    runtime.call("Status.Notify", level, message).catch(() => {});
  };

  const notifyError = (message, error) => {
    notify("error", `${message}：${messageOf(error)}`);
  };

  const isDark = (theme) => {
    if (theme === "dark") return true;
    if (theme === "light" || theme === "sepia") return false;
    return window.matchMedia("(prefers-color-scheme: dark)").matches;
  };

  const fontScale = (fontSize) => {
    if (fontSize === "large") return "1.3";
    if (fontSize === "medium") return "1.15";
    return "1";
  };

  const applyAppearance = (appearance) => {
    if (!appearance) return;
    state.prefs.theme = appearance.theme;
    state.prefs.fontSize = appearance.fontSize;
    applyAppearanceFromPrefs();
  };

  const applyAppearanceFromPrefs = () => {
    document.documentElement.classList.toggle("dark", isDark(state.prefs.theme));
    document.documentElement.classList.toggle("theme-sepia", state.prefs.theme === "sepia");
    document.documentElement.style.setProperty(
      "--app-text-scale",
      fontScale(state.prefs.fontSize),
    );
  };

  const applyPrefs = (prefs) => {
    if (!prefs) return;
    state.prefs = prefs;
    applyAppearanceFromPrefs();
  };

  // Article stack: each open article is one absolutely positioned layer so
  // the covered article keeps its DOM node (and its scroll position) while
  // the incoming one slides over or beneath it.
  const stack = {
    layers: [],
    swap: 0,
  };

  /** How an article layer behaves while the stack swaps. Entering rises
   * 24px while fading in; cover-down sheets down over the old article from
   * above; exit-up slides up and away, fully opaque, revealing the incoming
   * one beneath. */
  const kindFrames = {
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
   * frame, so the pre-animation frame cannot flash the end state. */
  const kindPins = {
    enter: { opacity: "0", transform: "translateY(24px)" },
    "cover-down": { opacity: "1", transform: "translateY(-100%)" },
    "exit-up": { opacity: "1", transform: "translateY(0)" },
  };

  const syncPane = () => {
    const hasLayers = stack.layers.length > 0;
    elements.empty.hidden = hasLayers;
    elements.article.hidden = !hasLayers;
  };

  const topLayer = () => stack.layers[stack.layers.length - 1] || null;

  /** Stop any in-flight slide and clear the pinned first frame. */
  const stopLayerAnimation = (layer) => {
    if (layer.animation) {
      layer.animation.cancel();
      layer.animation = null;
    }
    layer.element.style.opacity = "";
    layer.element.style.transform = "";
  };

  /**
   * Play a layer's slide for its current kind. The animation starts only
   * once the element is attached: an `animate` call made while the element
   * is still detached leaves the animation permanently pending (no start
   * time, `finished` never resolves) in both WebKit and Chromium, which
   * used to leave the pane black. Deferring to the next frame fixes that;
   * the first keyframe is pinned inline until then. An inline
   * `animation: none` keeps CSS animations from compositing over the slide
   * — WebKit lets a concurrent CSS animation win, which would cancel it.
   * Only the Web Animations API drives these slides; there is no CSS
   * animation on layers.
   */
  const playKind = (layer, kind, onFinished) => {
    stopLayerAnimation(layer);
    if (kind === "base") return;
    layer.element.style.animation = "none";
    layer.element.style.opacity = kindPins[kind].opacity;
    layer.element.style.transform = kindPins[kind].transform;
    requestAnimationFrame(() => {
      if (!layer.element.isConnected) return;
      layer.element.style.opacity = "";
      layer.element.style.transform = "";
      const animation = layer.element.animate(kindFrames[kind], {
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

  const setLayerKind = (layer, kind, onFinished) => {
    layer.element.dataset.kind = kind;
    playKind(layer, kind, onFinished);
  };

  /** Drop a leftover layer of one swap once its slide has finished. */
  const retireLayer = (layer, swap) => {
    if (!layer.retired || layer.swap !== swap) return;
    stopLayerAnimation(layer);
    const index = stack.layers.indexOf(layer);
    if (index >= 0) stack.layers.splice(index, 1);
    layer.element.remove();
    syncPane();
  };

  /** Backstop for engines that never report the slide as finished. */
  const scheduleRetire = (layer, swap) => {
    setTimeout(() => retireLayer(layer, swap), swapCleanupMs);
  };

  const createSkeleton = () => {
    const holder = document.createElement("div");
    holder.className = "article-skeleton";
    for (const className of ["loading-line", "loading-line short", "loading-line shorter"]) {
      const line = document.createElement("span");
      line.className = className;
      holder.append(line);
    }
    return holder;
  };

  const createLayer = (selection, kind, swap) => {
    const element = document.createElement("div");
    element.className = "article-layer";
    const body = document.createElement("div");
    body.className = "post-body";
    body.append(createSkeleton());
    element.append(body);
    const layer = { selection, element, body, swap, retired: false, animation: null };
    elements.article.append(element);
    stack.layers.push(layer);
    syncPane();
    setLayerKind(layer, kind);
    return layer;
  };

  const clearPost = () => {
    state.post = null;
    state.request += 1;
    cancelSnapshot();
    for (const layer of stack.layers) {
      stopLayerAnimation(layer);
      layer.element.remove();
    }
    stack.layers = [];
    syncPane();
  };

  const formatDateTime = (value) => {
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

  const externalHref = (href) => {
    if (!href) return null;
    try {
      const url = new URL(href);
      if (url.protocol !== "http:" && url.protocol !== "https:") return null;
      return url.toString();
    } catch {
      return null;
    }
  };

  const openLink = async (href) => {
    const url = externalHref(href);
    if (!url) return;
    try {
      await runtime.call("Settings.OpenLink", url);
    } catch (error) {
      notifyError("打开链接失败", error);
    }
  };

  const markRead = async (post) => {
    try {
      await runtime.call("Posts.MarkRead", post.feedId, post.guid, true);
    } catch (error) {
      notifyError("标记已读失败", error);
    }
  };

  const samePost = (first, second) =>
    first.feedId === second.feedId && first.guid === second.guid;

  /**
   * Swap the open article. The direction tells how the new selection sits
   * relative to the previous one in the native list: 1 later (the current
   * article slides up and away, revealing the incoming one beneath), -1
   * earlier (the new article sheets down over the current one), 0 unknown
   * (the first article fades in; a covered article with a missing
   * direction falls back to being covered). The entering layer mounts
   * immediately and slides while its body still loads, so quick keyboard
   * runs keep animating instead of waiting on `Posts.Get`.
   */
  const openPost = (selection, direction, markReadOnOpen) => {
    cancelSnapshot();
    const requestId = ++state.request;
    const covered = topLayer();

    if (covered && samePost(covered.selection, selection)) {
      return;
    }

    const swap = ++stack.swap;
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

    const entering = createLayer(selection, direction > 0 ? "enter" : "cover-down", swap);
    void loadIntoLayer(entering, requestId, markReadOnOpen);
  };

  const loadIntoLayer = async (layer, requestId, markReadOnOpen) => {
    try {
      const post = await runtime.call("Posts.Get", layer.selection.feedId, layer.selection.guid);
      if (requestId !== state.request) return;
      if (!stack.layers.includes(layer)) return;
      layer.body.innerHTML = post.content || post.summary;
      state.post = post;
      if (markReadOnOpen && !post.read) await markRead(post);
    } catch (error) {
      if (requestId !== state.request) return;
      restoreCovered(layer);
      notifyError("无法打开文章", error);
    }
  };

  /** Loading failed: drop the entering layer and bring the covered one back. */
  const restoreCovered = (layer) => {
    const index = stack.layers.indexOf(layer);
    if (index >= 0) stack.layers.splice(index, 1);
    stopLayerAnimation(layer);
    layer.element.remove();
    const previous = topLayer();
    if (previous && previous.retired) {
      previous.retired = false;
      previous.swap = 0;
      delete previous.element.dataset.kind;
      stopLayerAnimation(previous);
    }
    syncPane();
  };

  const scrollOrNext = () => {
    const layer = topLayer();
    if (!layer || !state.post) {
      runtime.call("Posts.SelectNeighbour", 1, false).catch(() => {});
      return;
    }
    const box = layer.element;
    if (box.clientHeight === 0) return;
    const bottomGap = box.scrollHeight - box.clientHeight - box.scrollTop;
    if (bottomGap > 1) {
      box.scrollBy({ top: box.clientHeight * 0.9, behavior: "smooth" });
      return;
    }
    runtime.call("Posts.SelectNeighbour", 1, false).catch(() => {});
  };

  const feedTitle = (feedId) => {
    const feed = state.feeds.find((item) => item.id === feedId);
    return feed?.title || "";
  };

  const closeSnapshot = () => {
    snapshot.request += 1;
    snapshot.busy = false;
    elements.snapshot.hidden = true;
    elements.snapshotProgress.hidden = true;
  };

  const cancelSnapshot = () => {
    if (!elements.snapshot.hidden || snapshot.busy) closeSnapshot();
  };

  const showSnapshotProgress = () => {
    snapshot.busy = true;
    elements.snapshot.hidden = false;
    elements.snapshotProgress.hidden = false;
    elements.snapshotProgressText.textContent = "正在准备文章快照…";
  };

  const waitForPaint = () => new Promise((resolve) => requestAnimationFrame(resolve));

  const waitForImages = async (card) => {
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

  const captureOffsets = (contentHeight, tileHeight) => {
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

  const buildSnapshotCard = (post, title) => {
    const card = document.createElement("article");
    card.className = "snapshot-card";

    const header = document.createElement("header");
    const meta = document.createElement("p");
    meta.className = "snapshot-meta";
    meta.textContent = [title.trim(), formatDateTime(post.publishedAt)].filter(Boolean).join(" · ");
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

    const body = document.createElement("div");
    body.className = "post-body";
    body.innerHTML = post.content || post.summary;

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

  const generateSnapshot = async (post, title, requestId) => {
    if (!post.content && !post.summary) {
      throw new Error("这篇文章没有可以生成快照的内容");
    }
    const viewportWidth = Math.max(1, Math.round(window.innerWidth));
    if (viewportWidth < minSnapshotWidth) {
      throw new Error("窗口太窄，无法生成快照");
    }
    const cardWidth = Math.min(maxSnapshotWidth, viewportWidth);

    const holder = document.createElement("div");
    holder.setAttribute("aria-hidden", "true");
    holder.className = "snapshot-overlay";
    const scroller = document.createElement("div");
    scroller.className = "snapshot-scroll";
    scroller.style.width = `${cardWidth}px`;
    const card = buildSnapshotCard(post, title);
    scroller.append(card);
    holder.append(scroller);
    document.body.append(holder);

    const ensureActive = () => {
      if (requestId !== snapshot.request) {
        throw new Error("snapshot-cancelled");
      }
    };

    try {
      ensureActive();
      await waitForPaint();
      await waitForImages(card);
      await waitForPaint();
      await waitForPaint();
      ensureActive();

      const contentHeight = scroller.scrollHeight;
      const tileHeight = scroller.clientHeight;
      if (contentHeight <= 0 || tileHeight <= 0) {
        throw new Error("无法准备文章快照");
      }
      const geometry = {
        viewportWidth,
        viewportHeight: tileHeight,
        cardWidth,
        contentHeight,
      };
      const offsets = captureOffsets(contentHeight, tileHeight);
      const tiles = [];
      let completed = 0;
      for (const offset of offsets) {
        ensureActive();
        scroller.scrollTop = offset;
        await waitForPaint();
        await waitForPaint();

        const actualOffset = Math.round(scroller.scrollTop);
        if (Math.abs(scroller.scrollHeight - contentHeight) > 1) {
          throw new Error("文章在生成快照时发生了变化");
        }
        const png = await runtime.call("Posts.CaptureSnapshot", {
          offset: actualOffset,
          geometry,
        });
        tiles.push({ offset: actualOffset, png, geometry });
        completed += 1;
        elements.snapshotProgressText.textContent =
          `正在生成文章快照… ${completed}/${offsets.length}`;
      }
      ensureActive();
      return tiles;
    } finally {
      holder.remove();
    }
  };

  const openSnapshot = async () => {
    const post = state.post;
    if (!post || snapshot.busy) return;
    if (!post.content && !post.summary) {
      notify("warning", "这篇文章没有可以生成快照的内容");
      return;
    }

    const requestId = ++snapshot.request;
    showSnapshotProgress();
    try {
      const tiles = await generateSnapshot(post, feedTitle(post.feedId), requestId);
      if (requestId !== snapshot.request) return;
      await runtime.call("Posts.StitchSnapshot", post.title || "", tiles);
      if (requestId !== snapshot.request) return;
      closeSnapshot();
    } catch (error) {
      if (requestId !== snapshot.request || messageOf(error) === "snapshot-cancelled") return;
      closeSnapshot();
      notifyError("无法生成快照", error);
    } finally {
      if (requestId === snapshot.request) snapshot.busy = false;
    }
  };

  const toggleRead = () => {
    const post = state.post;
    if (!post) return;
    runtime.call("Posts.MarkRead", post.feedId, post.guid, !post.read).catch((error) => {
      notifyError("更新阅读状态失败", error);
    });
  };

  const selectNeighbourFeed = (delta) => {
    runtime.call("Status.SelectNeighbourFeed", delta).catch((error) => {
      notifyError("无法切换订阅源", error);
    });
  };

  const showHelp = () => {
    runtime.call("Status.ShowHelp").catch((error) => {
      notifyError("无法打开快捷键帮助", error);
    });
  };

  const handleKey = (event) => {
    if (event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return;
    if (!elements.snapshot.hidden) {
      if (!snapshot.busy && event.key === "Escape") {
        event.preventDefault();
        closeSnapshot();
      }
      return;
    }

    switch (event.key) {
      case " ":
        event.preventDefault();
        scrollOrNext();
        break;
      case "m":
        toggleRead();
        break;
      case "v":
        if (state.post?.url) void openLink(state.post.url);
        break;
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

  const handleBodyClick = (event) => {
    const target = event.target instanceof Element ? event.target : null;
    if (!target) return;

    const image = target.closest("img");
    if (image && elements.article.contains(image)) {
      event.preventDefault();
      const source = image.currentSrc || image.getAttribute("src") || "";
      const alt = image.getAttribute("alt") || "";
      runtime.call("Posts.OpenImage", source, alt).catch((error) => {
        notifyError("无法打开图片", error);
      });
      return;
    }

    const anchor = target.closest("a");
    if (!anchor || !elements.article.contains(anchor)) return;
    const href = anchor.getAttribute("href");
    if (!href) return;
    event.preventDefault();
    void openLink(href);
  };

  const replaceFeeds = (snapshot) => {
    if (snapshot) state.feeds = snapshot.feeds;
  };

  const handleEvent = (event) => {
    if (event.revision <= state.revision) return;
    state.revision = event.revision;

    switch (event.eid) {
      case "state.snapshot":
        replaceFeeds(event.feeds);
        if (event.appearance) applyAppearance(event.appearance);
        break;
      case "feeds.changed":
        replaceFeeds(event.feeds);
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
          openPost(event.postSelection, event.direction || 0, state.prefs.markReadOnOpen);
        }
        break;
      case "post.updated":
        if (
          state.post &&
          event.post &&
          state.post.feedId === event.post.feedId &&
          state.post.guid === event.post.guid
        ) {
          Object.assign(state.post, event.post);
        }
        break;
      case "ui.command":
        handleCommand(event.command);
        break;
    }
  };

  const handleCommand = (command) => {
    switch (command) {
      case "add-feed":
        runtime.call("Feeds.OpenAddFeed").catch((error) => {
          notifyError("无法打开添加订阅源", error);
        });
        break;
      case "refresh":
        runtime.call("Feeds.RefreshAll").catch((error) => {
          notifyError("刷新订阅源失败", error);
        });
        break;
      case "snapshot":
        void openSnapshot();
        break;
      case "settings":
        runtime.call("Settings.OpenSettings").catch((error) => {
          notifyError("无法打开设置", error);
        });
        break;
    }
  };

  const start = async () => {
    document.documentElement.style.setProperty(
      "--mygo-text-scale",
      String(runtime.preferences.textScale),
    );
    if (runtime.preferences.accent) {
      document.documentElement.style.setProperty("--mygo-accent", runtime.preferences.accent);
    }
    applyAppearanceFromPrefs();
    const unsubscribe = runtime.on("app:event", handleEvent);
    const stopThemeWatch = () => {
      mediaQuery.removeEventListener("change", onSystemThemeChange);
    };
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const onSystemThemeChange = () => {
      if (state.prefs.theme === "system") applyAppearanceFromPrefs();
    };
    mediaQuery.addEventListener("change", onSystemThemeChange);
    window.addEventListener("keydown", handleKey);
    elements.article.addEventListener("click", handleBodyClick);
    window.addEventListener("unload", () => {
      unsubscribe();
      stopThemeWatch();
      window.removeEventListener("keydown", handleKey);
    });

    try {
      applyPrefs(await runtime.call("Settings.Get"));
      await runtime.call("Status.PageReady");
    } catch (error) {
      notifyError("无法读取本地数据", error);
    }
  };

  void start();
})();
