import { createSignal, onMount } from "solid-js";
import { Portal } from "solid-js/web";

/** How long the source-to-center preview animation takes, in milliseconds. */
const PREVIEW_MS = 240;

export type PreviewImage = {
  src: string;
  alt: string;
  rect: DOMRect;
  source: HTMLImageElement;
};

type ImagePreviewProps = {
  image: PreviewImage;
  /** Called after the closing animation finishes. */
  onClosed: () => void;
};

/** The centered size of one preview, clamped to the app viewport. */
function previewSize(image: PreviewImage): { width: number; height: number } {
  const naturalWidth = image.source.naturalWidth || image.rect.width;
  const naturalHeight = image.source.naturalHeight || image.rect.height;
  const maxWidth = Math.max(280, window.innerWidth * 0.88);
  const maxHeight = Math.max(200, window.innerHeight * 0.82);
  const fitScale = Math.min(maxWidth / naturalWidth, maxHeight / naturalHeight);
  const articleScale = image.rect.width > 0 ? image.rect.width / naturalWidth : 1;
  const scale = Math.min(fitScale, Math.max(articleScale, Math.min(3, fitScale)));

  return { width: naturalWidth * scale, height: naturalHeight * scale };
}

/** The transform that visually places a centered preview over the source. */
function sourceTransform(rect: DOMRect, previewWidth: number): string {
  const viewportCenterX = window.innerWidth / 2;
  const viewportCenterY = window.innerHeight / 2;
  const offsetX = rect.left + rect.width / 2 - viewportCenterX;
  const offsetY = rect.top + rect.height / 2 - viewportCenterY;
  return `translate(-50%, -50%) translate(${offsetX}px, ${offsetY}px) scale(${rect.width / previewWidth})`;
}

/** A lightbox that animates one article image from its position to center. */
export default function ImagePreview(props: ImagePreviewProps) {
  let overlayRef: HTMLDialogElement | undefined;
  let imageRef: HTMLImageElement | undefined;
  const [closing, setClosing] = createSignal(false);
  const size = previewSize(props.image);

  /** Plays the reverse FLIP animation before removing the preview. */
  const close = () => {
    if (closing()) return;
    setClosing(true);

    if (!overlayRef || !imageRef || typeof imageRef.animate !== "function") {
      props.onClosed();
      return;
    }

    const source = props.image.source;
    const rect =
      source.isConnected && source.getBoundingClientRect().width > 0
        ? source.getBoundingClientRect()
        : props.image.rect;
    const previewWidth = imageRef.getBoundingClientRect().width || size.width;
    const overlay = overlayRef;
    const previewImage = imageRef;
    overlay.style.opacity = "1";
    imageRef.style.transform = "translate(-50%, -50%)";
    requestAnimationFrame(() => {
      if (!overlay.isConnected || !previewImage.isConnected) return;
      const animation = previewImage.animate(
        [
          { transform: "translate(-50%, -50%)" },
          { transform: sourceTransform(rect, previewWidth) },
        ] satisfies Keyframe[],
        { duration: PREVIEW_MS, easing: "ease-out" },
      );
      overlay.animate([{ opacity: "1" }, { opacity: "0" }] satisfies Keyframe[], {
        duration: PREVIEW_MS,
        easing: "ease-out",
      });
      animation.finished.then(props.onClosed, props.onClosed);
    });
  };

  const bindImage = (element: HTMLImageElement) => {
    imageRef = element;
  };

  const bindOverlay = (element: HTMLDialogElement) => {
    overlayRef = element;
    element.style.opacity = "0";
    if (imageRef) imageRef.style.transform = sourceTransform(props.image.rect, size.width);
  };

  onMount(() => {
    const element = overlayRef;
    if (!element) return;
    if (typeof element.showModal === "function") element.showModal();
    if (typeof element.animate !== "function") {
      element.focus();
      return;
    }

    if (imageRef) imageRef.style.transform = sourceTransform(props.image.rect, size.width);
    requestAnimationFrame(() => {
      if (!element.isConnected || !imageRef?.isConnected) return;
      element.animate([{ opacity: "0" }, { opacity: "1" }] satisfies Keyframe[], {
        duration: PREVIEW_MS,
        easing: "ease-out",
      });
      const animation = imageRef.animate(
        [
          { transform: sourceTransform(props.image.rect, size.width) },
          { transform: "translate(-50%, -50%)" },
        ] satisfies Keyframe[],
        { duration: PREVIEW_MS, easing: "ease-out" },
      );
      animation.finished.then(
        () => {
          element.style.opacity = "";
          if (imageRef) imageRef.style.transform = "translate(-50%, -50%)";
        },
        () => {},
      );
      element.focus();
    });
  });

  return (
    <Portal>
      <dialog
        ref={bindOverlay}
        aria-modal="true"
        aria-label={props.image.alt || "图片预览"}
        class="image-preview-overlay fixed inset-0 z-[9000] flex h-dvh max-h-none w-dvw max-w-none items-center justify-center"
        data-image-preview=""
        data-closing={closing() ? "true" : undefined}
        onCancel={(event) => {
          event.preventDefault();
          close();
        }}
      >
        <button
          type="button"
          aria-hidden="true"
          tabindex="-1"
          class="absolute inset-0 cursor-zoom-out"
          onClick={close}
        />
        <img
          ref={bindImage}
          src={props.image.src}
          alt={props.image.alt}
          class="fixed top-1/2 left-1/2 max-h-[82vh] max-w-[88vw] rounded-md object-contain shadow-lg select-none"
          style={{ width: `${size.width}px`, height: `${size.height}px` }}
        />
      </dialog>
    </Portal>
  );
}
