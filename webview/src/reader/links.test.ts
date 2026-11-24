import { describe, expect, it } from "vitest";

import { externalHref } from "./links";

describe("externalHref", () => {
  it("keeps absolute http and https addresses", () => {
    expect(externalHref("https://example.com/post?a=1")).toBe("https://example.com/post?a=1");
    expect(externalHref("http://example.com/")).toBe("http://example.com/");
  });

  it("rejects non-web protocols", () => {
    expect(externalHref("javascript:alert(1)")).toBeNull();
    expect(externalHref("file:///etc/passwd")).toBeNull();
    expect(externalHref("ftp://example.com/x")).toBeNull();
    expect(externalHref("data:text/html,<p>x</p>")).toBeNull();
  });

  it("rejects relative and malformed addresses", () => {
    expect(externalHref("/relative/path")).toBeNull();
    expect(externalHref("not a url")).toBeNull();
    expect(externalHref("")).toBeNull();
    expect(externalHref(null)).toBeNull();
    expect(externalHref(undefined)).toBeNull();
  });
});
