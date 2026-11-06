export type Theme = "system" | "light" | "dark";

const darkQuery = "(prefers-color-scheme: dark)";

/** Whether a theme asks for the dark appearance right now. */
export function isDark(theme: string): boolean {
  if (theme === "dark") return true;
  if (theme === "light") return false;
  return window.matchMedia(darkQuery).matches;
}

/** Applies a theme to the page, as the `dark` class its CSS keys on. */
export function applyTheme(theme: string): void {
  document.documentElement.classList.toggle("dark", isDark(theme));
}

/** Calls listener when the system appearance changes. */
export function onSystemThemeChange(listener: () => void): () => void {
  const query = window.matchMedia(darkQuery);
  query.addEventListener("change", listener);
  return () => query.removeEventListener("change", listener);
}
