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
import { app, hideSidebar, keepSidebar, requestSidebar, savePaneSizes, start } from "~/stores/app";
import { subscribe } from "~/stores/events";

function collapsedPaneSizes(): [number, number] {
  const [sidebar, list] = app.prefs.paneSizes;
  const listSize = Math.min(Math.max(list / (100 - sidebar), 0.18), 0.55);
  return [listSize, 1 - listSize];
}

function dockedPaneSizes(): [number, number, number] {
  const sidebar = Math.min(Math.max(app.prefs.paneSizes[0] / 100, 0.12), 0.4);
  const list = Math.min(
    Math.max(app.prefs.paneSizes[1] / 100, 0.18),
    Math.min(0.55, 1 - sidebar - 0.25),
  );
  return [sidebar, list, 1 - sidebar - list];
}

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
          <Show
            when={!app.sidebar.collapsed}
            fallback={
              <Resizable
                class="flex h-full"
                initialSizes={collapsedPaneSizes()}
                onSizesChange={(sizes) => savePaneSizes(sizes, { sidebarCollapsed: true })}
              >
                <ResizablePanel initialSize={collapsedPaneSizes()[0]} minSize={0.18} maxSize={0.55}>
                  <PostList />
                </ResizablePanel>

                <ResizableHandle />

                <ResizablePanel initialSize={collapsedPaneSizes()[1]} minSize={0.25}>
                  <PostView />
                </ResizablePanel>
              </Resizable>
            }
          >
            <Resizable
              class="flex h-full"
              initialSizes={dockedPaneSizes()}
              onSizesChange={savePaneSizes}
            >
              <ResizablePanel initialSize={dockedPaneSizes()[0]} minSize={0.12} maxSize={0.4}>
                <FeedSidebar />
              </ResizablePanel>

              <ResizableHandle />

              <ResizablePanel initialSize={dockedPaneSizes()[1]} minSize={0.18} maxSize={0.55}>
                <PostList />
              </ResizablePanel>

              <ResizableHandle />

              <ResizablePanel initialSize={dockedPaneSizes()[2]} minSize={0.25}>
                <PostView />
              </ResizablePanel>
            </Resizable>
          </Show>

          <Show when={app.sidebar.collapsed}>
            <div
              class="floating-sidebar absolute top-2 bottom-2 left-2 z-40 w-80 max-w-[calc(100%-1rem)] overflow-hidden rounded-xl border bg-background shadow-lg"
              data-visible={app.sidebar.visible ? "true" : "false"}
              onMouseEnter={keepSidebar}
              onMouseLeave={hideSidebar}
            >
              <FeedSidebar />
            </div>
          </Show>

          <Show when={app.sidebar.collapsed}>
            <button
              type="button"
              class="absolute inset-y-0 left-0 z-40 w-2 cursor-pointer opacity-0"
              aria-label="显示订阅源"
              onFocus={requestSidebar}
              onMouseEnter={requestSidebar}
              onMouseLeave={hideSidebar}
            />
          </Show>
        </div>
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
