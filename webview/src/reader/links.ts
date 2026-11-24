/** Filter an untrusted address down to an absolute http(s) URL, or null. */
export const externalHref = (href: string | null | undefined): string | null => {
  if (!href) return null;
  try {
    const url = new URL(href);
    if (url.protocol !== "http:" && url.protocol !== "https:") return null;
    return url.toString();
  } catch {
    return null;
  }
};
