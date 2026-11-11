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
    body: document.getElementById("post-body"),
    empty: document.getElementById("empty"),
    loading: document.getElementById("loading"),
    snapshot: document.getElementById("snapshot-backdrop"),
    snapshotProgress: document.getElementById("snapshot-progress"),
    snapshotProgressText: document.getElementById("snapshot-progress-text"),
  };

  const maxSnapshotWidth = 760;
  const minSnapshotWidth = 320;
  const maxSnapshotTiles = 32;

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

  const showPane = (pane) => {
    elements.empty.hidden = pane !== "empty";
    elements.loading.hidden = pane !== "loading";
    elements.article.hidden = pane !== "article";
  };

  const clearPost = () => {
    state.post = null;
    state.request += 1;
    cancelSnapshot();
    elements.body.replaceChildren();
    showPane("empty");
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

  const showPost = (post) => {
    state.post = post;
    elements.body.innerHTML = post.content || post.summary;
    showPane("article");
    elements.article.scrollTop = 0;
  };

  const loadPost = async (selection, markReadOnOpen) => {
    cancelSnapshot();
    const requestId = ++state.request;
    showPane("loading");
    try {
      const post = await runtime.call("Posts.Get", selection.feedId, selection.guid);
      if (requestId !== state.request) return;
      showPost(post);
      if (markReadOnOpen && !post.read) await markRead(post);
    } catch (error) {
      if (requestId !== state.request) return;
      showPane(state.post ? "article" : "empty");
      notifyError("无法打开文章", error);
    }
  };

  const scrollOrNext = () => {
    if (!state.post) {
      runtime.call("Posts.SelectNeighbour", 1, false).catch(() => {});
      return;
    }
    const box = elements.article;
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
    if (image && elements.body.contains(image)) {
      event.preventDefault();
      const source = image.currentSrc || image.getAttribute("src") || "";
      const alt = image.getAttribute("alt") || "";
      runtime.call("Posts.OpenImage", source, alt).catch((error) => {
        notifyError("无法打开图片", error);
      });
      return;
    }

    const anchor = target.closest("a");
    if (!anchor || !elements.body.contains(anchor)) return;
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
          void loadPost(event.postSelection, state.prefs.markReadOnOpen);
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
    elements.body.addEventListener("click", handleBodyClick);
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
