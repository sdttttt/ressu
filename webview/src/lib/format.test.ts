import { describe, expect, it } from "vitest";

import { formatDateTime, formatReadingTime, formatWordCount, formatRelative } from "~/lib/format";

const now = new Date(2026, 4, 20, 15, 30, 0);

describe("formatDateTime", () => {
  it("keeps the time", () => {
    const value = new Date(2026, 4, 20, 9, 5, 0).toISOString();
    const text = formatDateTime(value);
    expect(text).toContain("2026");
    expect(text).toMatch(/09:05|9:05/);
  });
});

describe("formatWordCount", () => {
  it("formats readable counts and hides empty ones", () => {
    expect(formatWordCount(1234)).toBe("1,234 字");
    expect(formatWordCount(0)).toBe("");
  });
});

describe("formatReadingTime", () => {
  it("estimates at least one minute", () => {
    expect(formatReadingTime(1)).toBe("约 1 分钟");
    expect(formatReadingTime(500)).toBe("约 2 分钟");
  });

  it("switches to hours for long posts", () => {
    expect(formatReadingTime(30_000)).toBe("约 2 小时");
  });

  it("hides empty posts", () => {
    expect(formatReadingTime(0)).toBe("");
  });
});

describe("formatRelative", () => {
  it("describes recent times in words", () => {
    expect(formatRelative(new Date(2026, 4, 20, 15, 29, 40).toISOString(), now)).toBe("刚刚");
    expect(formatRelative(new Date(2026, 4, 20, 15, 5, 0).toISOString(), now)).toBe("25 分钟前");
    expect(formatRelative(new Date(2026, 4, 20, 11, 30, 0).toISOString(), now)).toBe("4 小时前");
    expect(formatRelative(new Date(2026, 4, 18, 15, 30, 0).toISOString(), now)).toBe("2 天前");
  });

  it("falls back to a date for old posts", () => {
    expect(formatRelative(new Date(2025, 0, 1, 0, 0, 0).toISOString(), now)).toContain("2025");
  });

  it("switches to a date after 30 days", () => {
    const boundary = new Date(now);
    boundary.setDate(boundary.getDate() - 30);
    const older = new Date(now);
    older.setDate(older.getDate() - 31);

    expect(formatRelative(boundary.toISOString(), now)).not.toBe("30 天前");
    expect(formatRelative(older.toISOString(), now)).toContain("2026");
  });
});
