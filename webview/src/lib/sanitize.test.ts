import { describe, expect, it } from "vitest";

import { externalHref, plainText, sanitizeHTML } from "~/lib/sanitize";

describe("sanitizeHTML", () => {
  it("drops scripts and styles", () => {
    const html = sanitizeHTML(
      '<p>hello</p><script>alert(1)</script><style>p{color:red}</style><iframe src="x"></iframe>',
    );
    expect(html).toContain("hello");
    expect(html).not.toContain("script");
    expect(html).not.toContain("style");
    expect(html).not.toContain("iframe");
  });

  it("drops inline styles", () => {
    const html = sanitizeHTML('<div style="color:red">x</div>');
    expect(html).not.toContain("style");
    expect(html).toContain("x");
  });

  it("points links at absolute URLs and opens them safely", () => {
    const html = sanitizeHTML('<a href="/posts/2">next</a>', "https://example.com/posts/1");
    expect(html).toContain('href="https://example.com/posts/2"');
    expect(html).toContain('target="_blank"');
    expect(html).toContain("noopener");
  });

  it("loads images lazily from absolute URLs without forcing a size", () => {
    const html = sanitizeHTML(
      '<img src="cover.png" width="800" height="600">',
      "https://example.com/posts/1",
    );
    expect(html).toContain('src="https://example.com/posts/cover.png"');
    expect(html).toContain('loading="lazy"');
    expect(html).not.toContain("width=");
    expect(html).not.toContain("height=");
  });

  it("returns an empty string for empty input", () => {
    expect(sanitizeHTML("")).toBe("");
  });
});

describe("plainText", () => {
  it("keeps the text of markup on one line", () => {
    expect(plainText("<p>hello <strong>there</strong></p>")).toBe("hello there");
    expect(plainText("<p>a\n\n b</p>")).toBe("a b");
  });

  it("cuts long text", () => {
    const text = plainText(`<p>${"字".repeat(20)}</p>`, 5);
    expect(text).toBe("字字字字字…");
  });

  it("returns nothing for empty input", () => {
    expect(plainText("")).toBe("");
  });
});

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
