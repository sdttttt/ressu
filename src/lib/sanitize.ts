import DOMPurify from "dompurify";

const FORBID_TAGS = ["script", "style", "iframe", "object", "embed", "form", "input", "button"];
const FORBID_ATTR = ["style", "srcset", "onerror", "onload"];

function absolute(base: string, reference: string): string {
  if (!base || !reference) return reference;
  try {
    return new URL(reference, base).toString();
  } catch {
    return reference;
  }
}

/**
 * Cleans the HTML of a post so it can be shown in the app: scripts and styles
 * are dropped, links point at absolute URLs and open through the app, and
 * images load lazily.
 */
export function sanitizeHTML(html: string, baseURL = ""): string {
  if (!html) return "";

  const clean = DOMPurify.sanitize(html, {
    USE_PROFILES: { html: true },
    FORBID_TAGS,
    FORBID_ATTR,
    ADD_ATTR: ["target"],
  });

  const doc = new DOMParser().parseFromString(clean, "text/html");

  for (const link of doc.querySelectorAll("a[href]")) {
    const href = link.getAttribute("href") ?? "";
    link.setAttribute("href", absolute(baseURL, href));
    link.setAttribute("target", "_blank");
    link.setAttribute("rel", "noopener noreferrer");
  }

  for (const image of doc.querySelectorAll("img")) {
    const src = image.getAttribute("src");
    if (src) image.setAttribute("src", absolute(baseURL, src));
    image.setAttribute("loading", "lazy");
    image.setAttribute("decoding", "async");
    image.removeAttribute("width");
    image.removeAttribute("height");
  }

  for (const source of doc.querySelectorAll("source[src]")) {
    source.setAttribute("src", absolute(baseURL, source.getAttribute("src") ?? ""));
  }

  return doc.body.innerHTML;
}

/** Returns the text a piece of markup shows, on one line. */
export function plainText(html: string, limit = 280): string {
  if (!html) return "";
  const doc = new DOMParser().parseFromString(html, "text/html");
  const text = (doc.body.textContent ?? "").replace(/\s+/g, " ").trim();
  if (text.length <= limit) return text;
  return `${text.slice(0, limit).trimEnd()}…`;
}

/** Parses the URL a link points at, or null when it is the app's own page. */
export function externalHref(href: string): string | null {
  if (!href) return null;
  try {
    const url = new URL(href);
    if (url.protocol !== "http:" && url.protocol !== "https:") return null;
    return url.toString();
  } catch {
    return null;
  }
}
