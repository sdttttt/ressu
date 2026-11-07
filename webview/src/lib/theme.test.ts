import { afterEach, describe, expect, it } from "vitest";

import { applyFontSize, applyTheme, fontScale, isDark } from "~/lib/theme";

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
    expect(isDark("sepia")).toBe(false);
    expect(isDark("light")).toBe(false);
  });
});

describe("applyTheme", () => {
  afterEach(() => {
    document.documentElement.className = "";
  });

  it("keeps the light palette for sepia and only toggles its background", () => {
    applyTheme("sepia");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(document.documentElement.classList.contains("theme-sepia")).toBe(true);

    applyTheme("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    expect(document.documentElement.classList.contains("theme-sepia")).toBe(false);

    applyTheme("light");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
    expect(document.documentElement.classList.contains("theme-sepia")).toBe(false);
  });
});
