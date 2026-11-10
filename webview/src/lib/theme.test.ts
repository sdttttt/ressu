import { afterEach, describe, expect, it } from "vitest";

import { applyFontSize, applyTheme } from "~/lib/theme";

describe("applyFontSize", () => {
  afterEach(() => {
    document.documentElement.style.removeProperty("--app-text-scale");
  });

  it("scales the root font size through the app variable", () => {
    applyFontSize("medium");
    expect(document.documentElement.style.getPropertyValue("--app-text-scale")).toBe("1.15");

    applyFontSize("small");
    expect(document.documentElement.style.getPropertyValue("--app-text-scale")).toBe("1");
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
