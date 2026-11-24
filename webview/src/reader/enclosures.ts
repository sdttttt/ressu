/** A display name for an attachment: the decoded last path segment. */
export const enclosureName = (url: string): string => {
  try {
    const segments = new URL(url).pathname.split("/").filter(Boolean);
    const segment = segments.pop() || "attachment";
    try {
      return decodeURIComponent(segment);
    } catch {
      return segment;
    }
  } catch {
    return "attachment";
  }
};

/** A human size for an attachment, empty when the declared length is unusable. */
export const enclosureSize = (length: number): string => {
  if (typeof length !== "number" || !Number.isFinite(length) || length <= 0) {
    return "";
  }
  if (length >= 1 << 30) return `${(length / (1 << 30)).toFixed(1)} GB`;
  if (length >= 1 << 20) return `${(length / (1 << 20)).toFixed(1)} MB`;
  if (length >= 1 << 10) return `${Math.round(length / (1 << 10))} KB`;
  return `${length} B`;
};
