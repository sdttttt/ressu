(() => {
  "use strict";

  const runtime = globalThis.mygo;
  if (!runtime) {
    throw new Error("mygo runtime not found");
  }

  const elements = {
    stage: document.getElementById("stage"),
    veil: document.getElementById("veil"),
    hud: document.getElementById("hud"),
    hudText: document.getElementById("hud-text"),
    error: document.getElementById("error"),
    errorText: document.getElementById("error-text"),
    errorClose: document.getElementById("error-close"),
  };

  const maxSnapshotTiles = 32;

  const messageOf = (error) => {
    if (error instanceof Error && error.message) return error.message;
    return String(error ?? "未知错误");
  };

  const setBusy = (busy, text) => {
    elements.veil.hidden = !busy;
    elements.hud.hidden = !busy;
    if (text) elements.hudText.textContent = text;
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

  const buildCard = (post, feedTitle) => {
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

    // Post.Content is sanitized HTML from the database; nothing else may
    // enter the card as HTML.
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

  const captureTiles = async (post) => {
    if (!post.content && !post.summary) {
      throw new Error("这篇文章没有可以生成快照的内容");
    }

    setBusy(true, "正在准备文章快照…");
    await waitForPaint();
    await waitForImages(post.card);
    await waitForPaint();
    await waitForPaint();

    const viewportWidth = Math.max(1, Math.round(window.innerWidth));
    const cardWidth = Math.round(post.card.getBoundingClientRect().width);
    const cardLeft = Math.max(0, Math.round((viewportWidth - cardWidth) / 2));
    const contentHeight = elements.stage.scrollHeight;
    const tileHeight = elements.stage.clientHeight;
    if (contentHeight <= 0 || tileHeight <= 0 || cardWidth <= 0) {
      throw new Error("无法准备文章快照");
    }
    const geometry = {
      viewportWidth,
      viewportHeight: tileHeight,
      cardWidth,
      cardLeft,
      contentHeight,
    };
    const offsets = captureOffsets(contentHeight, tileHeight);
    const tiles = [];
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
      const png = await runtime.call("Posts.CaptureSnapshot", {
        offset: actualOffset,
        geometry,
      });
      tiles.push({ offset: actualOffset, png, geometry });
      completed += 1;
      setBusy(true, `正在生成文章快照… ${completed}/${offsets.length}`);
    }
    return tiles;
  };

  const fail = (error) => {
    setBusy(false);
    elements.errorText.textContent = messageOf(error);
    elements.error.hidden = false;
  };

  const run = async () => {
    try {
      const target = await runtime.call("Snapshot.Selected");
      if (!target || !target.guid) {
        throw new Error("没有选中的文章");
      }
      const post = await runtime.call("Posts.Get", target.feedId, target.guid);
      const card = buildCard(post, target.feedTitle || "");
      elements.stage.append(card);
      post.card = card;
      const tiles = await captureTiles(post);
      setBusy(true, "正在合成文章快照…");
      await runtime.call("Posts.StitchSnapshot", post.title || "", tiles);
      // A successful stitch opens the native preview and closes this window.
    } catch (error) {
      fail(error);
    }
  };

  const closeWindow = () => {
    runtime.call("Snapshot.Cancel").catch(() => {});
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
})();
