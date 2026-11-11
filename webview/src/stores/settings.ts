import { toast } from "~/lib/notifications";
import { externalHref } from "~/lib/sanitize";
import { Settings } from "~/mygo";
import { messageOf } from "~/stores/errors";

/** Opens a link outside the reader, the way the preferences ask for. */
export async function openLink(href: string): Promise<void> {
  const url = externalHref(href);
  if (!url) return;
  try {
    await Settings.openLink(url);
  } catch (error) {
    toast.error(messageOf(error));
  }
}

/** Opens the native settings window. */
export async function openSettings(): Promise<void> {
  try {
    await Settings.openSettings();
  } catch (error) {
    toast.error(messageOf(error));
  }
}
