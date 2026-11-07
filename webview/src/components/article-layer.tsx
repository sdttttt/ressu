import type { JSXElement } from "solid-js";
import { createEffect } from "solid-js";

import { cn } from "~/lib/utils";

/** How long an article layer's slide takes, in milliseconds. */
export const SLIDE_MS = 300;

/** How an article layer behaves inside the pane while the stack swaps. */
export type LayerKind = "base" | "enter" | "cover-down" | "exit-up";

/**
 * Layer keyframes, keyed by kind. Entering rises 24px while fading in;
 * cover-down sheets down over the old article from above; the exit-up layer
 * slides up and away, fully opaque, revealing the incoming one beneath.
 */
const KIND_FRAMES: Record<Exclude<LayerKind, "base">, Keyframe[]> = {
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

/** Static first-frame pins for kinds whose animation starts on a later frame. */
const KIND_PINS: Record<Exclude<LayerKind, "base">, { opacity: string; transform: string }> = {
  enter: { opacity: "0", transform: "translateY(24px)" },
  "cover-down": { opacity: "1", transform: "translateY(-100%)" },
  "exit-up": { opacity: "1", transform: "translateY(0)" },
};

/** Positioning and (CSS fallback) entry classes per kind. Every kind must be
 * absolutely positioned: the stack container clips its overflow, so a
 * statically positioned layer would grow past the pane and leave its inner
 * scroller unconstrained — the article's scrollbar would never appear. */
const KIND_CLASS: Record<LayerKind, string> = {
  base: "pointer-events-none absolute inset-0 z-0",
  enter: "absolute inset-0 animate-in duration-300 ease-out fade-in slide-in-from-bottom-6",
  "cover-down":
    "absolute inset-0 z-10 animate-in bg-background duration-300 ease-out slide-in-from-top-full",
  "exit-up": "absolute inset-0 z-10 bg-background",
};

/**
 * Play a layer's animation for its current kind. Called from a reactive
 * effect, so a reused DOM node re-animates when a swap flips its kind — that
 * is how the exit-up slide runs on a node that mounted long before. The
 * animation must start only once the element is attached: an `animate` call
 * made while the element is still detached leaves the animation permanently
 * pending (no start time, `ready` never resolves) in both WebKit and
 * Chromium, which used to leave the pane black. Deferring to the next frame
 * fixes that; the first keyframe is pinned inline until then so the
 * pre-animation frame cannot flash the end state. An inline
 * `animation: none` keeps the CSS entry classes from compositing over the
 * slide — WebKit lets a concurrent CSS animation win, which would cancel it.
 */
function playKind(
  ref: HTMLDivElement,
  kind: LayerKind,
  duration: number,
  onFinished?: () => void,
): void {
  if (kind === "base") return;
  if (typeof ref.animate !== "function") return; // non-browser environment (tests)

  // Kill any CSS animation before it can start; this does not touch WAAPI.
  ref.style.animation = "none";
  ref.style.opacity = KIND_PINS[kind].opacity;
  ref.style.transform = KIND_PINS[kind].transform;
  requestAnimationFrame(() => {
    if (!ref.isConnected) return;
    ref.style.opacity = "";
    ref.style.transform = "";
    const animation = ref.animate(KIND_FRAMES[kind], { duration, easing: "ease-out" });
    // A rejected `finished` means the layer was replaced mid-flight; the
    // newer swap rebuilds the stack anyway.
    if (onFinished) animation.finished.then(onFinished, () => {});
  });
}

type ArticleLayerProps = {
  kind: LayerKind;
  /** The swap this layer belongs to; reported back when its slide ends. */
  swap?: number;
  /** Called once this layer's slide has finished (never for "base"). */
  onFinished?: (swap: number) => void;
  children: JSXElement;
};

/**
 * One animated pane of the article stack. The layer content is rendered by
 * the parent as `children`; this component only positions the layer and runs
 * its slide. The element is keyed by the parent's layer object, so across
 * swaps the same node flips kind in place — children keep their state (images,
 * scroll position) and the kind effect replays the motion.
 */
export default function ArticleLayer(props: ArticleLayerProps) {
  let ref: HTMLDivElement | undefined;
  const notify = () => {
    if (props.swap !== undefined) props.onFinished?.(props.swap);
  };
  const bindLayer = (el: HTMLDivElement) => {
    ref = el;
    playKind(el, props.kind, SLIDE_MS, notify);
  };
  // The node is reused across swaps (a flip of props.kind), so the animation
  // tracks the kind reactively instead of running only on mount — that is
  // what makes the exit-up slide play on a node that mounted long before.
  createEffect(() => {
    const kind = props.kind;
    if (ref) playKind(ref, kind, SLIDE_MS, notify);
  });

  return (
    <div
      ref={bindLayer}
      data-layer={props.kind}
      class={cn("flex min-h-0 flex-1 flex-col", KIND_CLASS[props.kind])}
    >
      {props.children}
    </div>
  );
}
