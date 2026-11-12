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
    prefs: {
      theme: "system",
      fontSize: "medium",
    },
  };

  const elements = {
    article: document.getElementById("article"),
    empty: document.getElementById("empty"),
  };

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
    for (const layer of stack.layers) {
      stopLayerAnimation(layer);
      layer.element.remove();
    }
    stack.layers = [];
    syncPane();
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
  const openPost = (selection, direction) => {
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
    void loadIntoLayer(entering, requestId);
  };

  const loadIntoLayer = async (layer, requestId) => {
    try {
      const post = await runtime.call("Posts.Get", layer.selection.feedId, layer.selection.guid);
      if (requestId !== state.request) return;
      if (!stack.layers.includes(layer)) return;
      layer.body.innerHTML = post.content || post.summary;
      state.post = post;
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

  const toggleRead = () => {
    const post = state.post;
    if (!post) return;
    runtime.call("Posts.ToggleRead", post.feedId, post.guid).catch((error) => {
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

  const handleEvent = (event) => {
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
