import { Show, createEffect } from "solid-js";

import { Attachments } from "./attachments";
import { Layer, playKind } from "./store";
import type { LayerPost } from "./store";

/** The loading placeholder shown while an article body is on its way. */
function Skeleton() {
  return (
    <div class="article-skeleton">
      <span class="loading-line" />
      <span class="loading-line short" />
      <span class="loading-line shorter" />
    </div>
  );
}

/**
 * The article body. Post.Content is sanitized HTML and goes straight into
 * innerHTML; Post.Summary is plain text and must never be parsed as
 * markup. The wrapper keeps the content one child of .post-body so the
 * attachment bar stays a sibling, like the imperative version appended it.
 */
function PostContent(props: { post: LayerPost }) {
  let element!: HTMLDivElement;
  createEffect(
    () => props.post,
    (post) => {
      if (post.html) {
        element.innerHTML = post.html;
      } else {
        element.textContent = post.summary;
      }
    },
  );
  return <div class="post-content" ref={element} />;
}

/**
 * One open article: an absolutely positioned layer whose DOM node (and
 * scroll position) survives being covered by the next one. The layer's
 * slide is driven by an effect on its kind; the store decides the kinds.
 */
export function ArticleLayerView(props: { layer: Layer }) {
  const layer = props.layer;

  createEffect(
    () => layer.kind(),
    (kind) => {
      playKind(layer, kind);
      return () => {
        if (layer.animation) {
          layer.animation.cancel();
          layer.animation = null;
        }
      };
    },
  );

  return (
    <div
      class="article-layer"
      data-kind={layer.kind()}
      ref={(element) => {
        layer.element = element;
      }}
    >
      <div class="post-body">
        <Show when={layer.post()} fallback={<Skeleton />}>
          {(post) => (
            <>
              <PostContent post={post()} />
              <Attachments enclosures={post().enclosures} />
            </>
          )}
        </Show>
      </div>
    </div>
  );
}
