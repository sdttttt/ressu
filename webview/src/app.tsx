import { Rss } from "lucide-solid";
import { Show, onCleanup, onMount } from "solid-js";

import PostView from "~/components/post-view";
import ShortcutsDialog from "~/components/shortcuts-dialog";
import SnapshotDialog from "~/components/snapshot-dialog";
import { Skeleton } from "~/components/ui/skeleton";
import { installShortcuts, isMac } from "~/lib/commands";
import { applyTheme, onSystemThemeChange } from "~/lib/theme";
import { app, start } from "~/stores/app";
import { subscribe } from "~/stores/events";

export default function App() {
  onMount(() => {
    const unsubscribe = subscribe();
    const stopThemeWatch = onSystemThemeChange(() => {
      if (app.prefs.theme === "system") applyTheme("system");
    });
    // The system menu owns the mod-key shortcuts on macOS, so the page only
    // registers the bare ones there; other platforms get both.
    const stopShortcuts = installShortcuts({ skipMod: isMac });
    void start();

    onCleanup(() => {
      unsubscribe();
      stopThemeWatch();
      stopShortcuts?.();
    });
  });

  return (
    <div class="flex h-full flex-col bg-background text-foreground">
      <Show when={app.ready} fallback={<Splash />}>
        <div class="relative min-h-0 flex-1">
          <PostView />
        </div>
      </Show>

      <ShortcutsDialog />
      <SnapshotDialog />
    </div>
  );
}

function Splash() {
  return (
    <div class="flex min-h-0 flex-1 flex-col items-center justify-center gap-4">
      <Rss class="size-10 text-muted-foreground/60" />
      <div class="w-48 space-y-2">
        <Skeleton class="h-2.5 w-full" />
        <Skeleton class="h-2.5 w-4/5" />
      </div>
    </div>
  );
}
