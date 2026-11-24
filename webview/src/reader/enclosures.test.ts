import { describe, expect, it } from "vitest";

import { enclosureName, enclosureSize } from "./enclosures";

describe("enclosureName", () => {
  it("uses the decoded last path segment", () => {
    expect(enclosureName("https://cdn.example/episodes/ep01.mp3")).toBe("ep01.mp3");
    expect(enclosureName("https://cdn.example/%E7%AC%AC%E4%B8%80%E9%9B%86.mp3")).toBe(
      "第一集.mp3",
    );
  });

  it("falls back for root paths and invalid addresses", () => {
    expect(enclosureName("https://cdn.example/")).toBe("attachment");
    expect(enclosureName("not a url")).toBe("attachment");
  });
});

describe("enclosureSize", () => {
  it("formats bytes, KB, MB and GB", () => {
    expect(enclosureSize(512)).toBe("512 B");
    expect(enclosureSize(2048)).toBe("2 KB");
    expect(enclosureSize(3 * 1024 * 1024)).toBe("3.0 MB");
    expect(enclosureSize(1.5 * 1024 * 1024 * 1024)).toBe("1.5 GB");
  });

  it("is empty for unusable lengths", () => {
    expect(enclosureSize(0)).toBe("");
    expect(enclosureSize(-1)).toBe("");
    expect(enclosureSize(Number.NaN)).toBe("");
    expect(enclosureSize(Number.POSITIVE_INFINITY)).toBe("");
  });
});
