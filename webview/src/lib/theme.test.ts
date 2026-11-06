import { afterEach, describe, expect, it } from "vitest";

import { applyFontSize, fontScale, isDark } from "~/lib/theme";

describe("fontScale", () => {
  it("maps every preset", () => {
    expect(fontScale("small")).toBe(1);
    expect(fontScale("medium")).toBeGreaterThan(1);
    expect(fontScale("large")).toBeGreaterThan(fontScale("medium"));
  });

  it("falls back to the small scale", () => {
    expect(fontScale("huge")).toBe(1);
  });
});

describe("applyFontSize", () => {
  afterEach(() => {
    document.documentElement.style.removeProperty("--app-text-scale");
  });

  it("scales the root font size through the app variable", () => {
    applyFontSize("medium");
    expect(document.documentElement.style.getPropertyValue("--app-text-scale")).toBe(
      String(fontScale("medium")),
    );

    applyFontSize("small");
    expect(document.documentElement.style.getPropertyValue("--app-text-scale")).toBe("1");
  });
});

describe("isDark", () => {
  it("answers the explicit themes", () => {
    expect(isDark("dark")).toBe(true);
    expect(isDark("light")).toBe(false);
  });
});
