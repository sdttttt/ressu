import { Status } from "~/mygo";

type NotificationLevel = "error" | "warning" | "info" | "success";

function notify(level: NotificationLevel, message: string): void {
  try {
    void Promise.resolve(Status.notify(level, message)).catch(() => undefined);
  } catch {
    return;
  }
}

/** Sends a native application notification through the Go binding. */
export const toast = {
  error(message: string) {
    notify("error", message);
  },
  warning(message: string) {
    notify("warning", message);
  },
  info(message: string) {
    notify("info", message);
  },
  success(message: string) {
    notify("success", message);
  },
};
