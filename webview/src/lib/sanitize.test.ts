import { describe, expect, it } from "vitest";

import { externalHref } from "~/lib/sanitize";

describe("externalHref", () => {
  it("accepts web URLs", () => {
    expect(externalHref("https://example.com/a?b=1")).toBe("https://example.com/a?b=1");
    expect(externalHref("http://example.com/")).toBe("http://example.com/");
  });

  it("refuses anything else", () => {
    expect(externalHref("/relative")).toBeNull();
    expect(externalHref("javascript:alert(1)")).toBeNull();
    expect(externalHref("mailto:a@b.c")).toBeNull();
    expect(externalHref("")).toBeNull();
  });
});
