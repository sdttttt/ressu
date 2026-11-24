export const maxSnapshotTiles = 32;

/**
 * The scroll offsets one tile per viewport needs to cover the content,
 * each tile overlapping the previous by one pixel so stitches have no
 * seams. Throws when the article would need more tiles than the capture
 * window may take.
 */
export const captureOffsets = (contentHeight: number, tileHeight: number): number[] => {
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
