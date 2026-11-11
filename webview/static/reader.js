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
    feedId: 0,
    feeds: [],
    prefs: {
      theme: "system",
      fontSize: "medium",
      markReadOnOpen: true,
    },
  };

  const elements = {
    article: document.getElementById("article"),
    body: document.getElementById("post-body"),
    empty: document.getElementById("empty"),
    loading: document.getElementById("loading"),
    help: document.getElementById("help-backdrop"),
  };

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

  const selectNeighbourFeed = (delta) => {
    if (state.feeds.length === 0) return;
    const currentIndex = state.feeds.findIndex((feed) => feed.id === state.feedId);
    const fallbackIndex = delta === 1 ? 0 : state.feeds.length - 1;
    let index = currentIndex === -1 ? fallbackIndex : currentIndex + delta;
    index = Math.max(0, Math.min(index, state.feeds.length - 1));
    runtime.call("Status.SelectFeed", state.feeds[index].id).catch(() => {});
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

  const toggleRead = () => {
    const post = state.post;
    if (!post) return;
    runtime.call("Posts.MarkRead", post.feedId, post.guid, !post.read).catch((error) => {
      notifyError("更新阅读状态失败", error);
    });
  };

  const handleKey = (event) => {
    if (event.isComposing || event.ctrlKey || event.metaKey || event.altKey) return;
    if (!elements.help.hidden && event.key !== "Escape") return;

    switch (event.key) {
      case "?":
        event.preventDefault();
        elements.help.hidden = false;
        break;
      case "Escape":
        elements.help.hidden = true;
        break;
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
        selectNeighbourFeed(-1);
        break;
      case "l":
        selectNeighbourFeed(1);
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
        state.feedId = event.selection?.feedId ?? 0;
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
    elements.help.addEventListener("click", (event) => {
      if (event.target === elements.help) elements.help.hidden = true;
    });
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
