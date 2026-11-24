import { describe, expect, it } from "vitest";

import { captureOffsets, maxSnapshotTiles } from "./offsets";

describe("captureOffsets", () => {
  it("needs one tile when the content fits the viewport", () => {
    expect(captureOffsets(600, 800)).toEqual([0]);
    expect(captureOffsets(800, 800)).toEqual([0]);
  });

  it("overlaps tiles by one pixel and clamps the last one", () => {
    const offsets = captureOffsets(2000, 800);
    expect(offsets[0]).toBe(0);
    for (let i = 1; i < offsets.length; i++) {
      expect(offsets[i] - offsets[i - 1]).toBeLessThanOrEqual(799);
      expect(offsets[i]).toBeGreaterThan(offsets[i - 1]);
    }
    expect(offsets[offsets.length - 1]).toBe(2000 - 800);
  });

  it("throws when the article needs more tiles than allowed", () => {
    expect(() => captureOffsets((maxSnapshotTiles + 1) * 800, 800)).toThrow(
      "文章太长，无法生成一张快照",
    );
  });
});
