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
