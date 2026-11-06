const timeFormat = new Intl.DateTimeFormat("zh-CN", {
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

const monthDayFormat = new Intl.DateTimeFormat("zh-CN", { month: "long", day: "numeric" });

const dateFormat = new Intl.DateTimeFormat("zh-CN", {
  year: "numeric",
  month: "long",
  day: "numeric",
});

const dateTimeFormat = new Intl.DateTimeFormat("zh-CN", {
  year: "numeric",
  month: "long",
  day: "numeric",
  hour: "2-digit",
  minute: "2-digit",
  hour12: false,
});

function toDate(value: string): Date | null {
  if (!value) return null;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? null : date;
}

/** Formats a post's date the way a list shows it: a time today, else a date. */
export function formatPostDate(value: string, now: Date = new Date()): string {
  const date = toDate(value);
  if (!date) return "";

  const sameDay =
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate();
  if (sameDay) return timeFormat.format(date);

  if (date.getFullYear() === now.getFullYear()) return monthDayFormat.format(date);
  return dateFormat.format(date);
}

/** Formats a post's date with its time, as the reader shows it. */
export function formatDateTime(value: string): string {
  const date = toDate(value);
  return date ? dateTimeFormat.format(date) : "";
}

/** Formats how long ago something happened, in words. */
export function formatRelative(value: string, now: Date = new Date()): string {
  const date = toDate(value);
  if (!date) return "";

  const minutes = Math.round((now.getTime() - date.getTime()) / 60_000);
  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${minutes} 分钟前`;

  const hours = Math.round(minutes / 60);
  if (hours < 24) return `${hours} 小时前`;

  const days = Math.round(hours / 24);
  if (days < 30) return `${days} 天前`;

  return dateFormat.format(date);
}

/** Cuts text to a length, ending it with an ellipsis. */
export function truncate(value: string, limit: number): string {
  const text = value.trim();
  if (text.length <= limit) return text;
  return `${text.slice(0, limit).trimEnd()}…`;
}
