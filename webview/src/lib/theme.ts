export type Theme = "system" | "light" | "dark";

const darkQuery = "(prefers-color-scheme: dark)";

/** Whether a theme asks for the dark appearance right now. */
export function isDark(theme: string): boolean {
  if (theme === "dark") return true;
  if (theme === "light") return false;
  return window.matchMedia(darkQuery).matches;
}

/** The scale factors of the font size preference. */
const fontScales: Record<string, number> = {
  small: 1,
  medium: 1.15,
  large: 1.3,
};

/** The scale factor of a font size preference, defaulting to small. */
export function fontScale(fontSize: string): number {
  return fontScales[fontSize] ?? 1;
}

/** Applies a theme to the page, as the `dark` class its CSS keys on. */
export function applyTheme(theme: string): void {
  document.documentElement.classList.toggle("dark", isDark(theme));
}

/** Applies the font size to the page, scaling the root font size with it. */
export function applyFontSize(fontSize: string): void {
  document.documentElement.style.setProperty("--app-text-scale", String(fontScale(fontSize)));
}

/** Calls listener when the system appearance changes. */
export function onSystemThemeChange(listener: () => void): () => void {
  const query = window.matchMedia(darkQuery);
  query.addEventListener("change", listener);
  return () => query.removeEventListener("change", listener);
}
