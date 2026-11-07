import { Rss } from "lucide-solid";
import { Show, onCleanup, onMount } from "solid-js";

import AddFeedDialog from "~/components/add-feed-dialog";
import FeedSidebar from "~/components/feed-sidebar";
import PostList from "~/components/post-list";
import PostView from "~/components/post-view";
import SettingsDialog from "~/components/settings-dialog";
import ShortcutsDialog from "~/components/shortcuts-dialog";
import SnapshotDialog from "~/components/snapshot-dialog";
import StatusBar from "~/components/status-bar";
import { Resizable, ResizableHandle, ResizablePanel } from "~/components/ui/resizable";
import { Skeleton } from "~/components/ui/skeleton";
import { Toaster } from "~/components/ui/sonner";
import { installShortcuts, isMac } from "~/lib/commands";
import { applyTheme, onSystemThemeChange } from "~/lib/theme";
import { app, savePaneSizes, start } from "~/stores/app";
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
        <Resizable
          class="flex min-h-0 flex-1"
          initialSizes={app.prefs.paneSizes.map((size) => size / 100)}
          onSizesChange={savePaneSizes}
        >
          <ResizablePanel initialSize={app.prefs.paneSizes[0] / 100} minSize={0.12} maxSize={0.4}>
            <FeedSidebar />
          </ResizablePanel>

          <ResizableHandle />

          <ResizablePanel initialSize={app.prefs.paneSizes[1] / 100} minSize={0.18} maxSize={0.55}>
            <PostList />
          </ResizablePanel>

          <ResizableHandle />

          <ResizablePanel minSize={0.25}>
            <PostView />
          </ResizablePanel>
        </Resizable>
      </Show>

      <StatusBar />
      <AddFeedDialog />
      <SettingsDialog />
      <ShortcutsDialog />
      <SnapshotDialog />
      <Toaster position="bottom-right" />
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
