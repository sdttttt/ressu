import { Loader2 } from "lucide-solid";
import { Show } from "solid-js";

import { formatRelative } from "~/lib/format";
import { app, currentFeed, totalUnread } from "~/stores/app";

export default function StatusBar() {
  return (
    <footer class="flex h-7 shrink-0 items-center gap-2 border-t bg-background px-3 text-[11px] text-muted-foreground">
      <Show
        when={app.sync.active}
        fallback={
          <span class="truncate tabular-nums">
            {app.feeds.length} 个订阅源 · {totalUnread()} 篇未读
          </span>
        }
      >
        <Loader2 class="size-3 shrink-0 animate-spin" />
        <span class="truncate tabular-nums">
          正在刷新 {app.sync.done}/{app.sync.total}
          <Show when={app.sync.feed}> · {app.sync.feed}</Show>
        </span>
      </Show>

      <div class="flex-1" />

      <Show when={app.sync.error}>
        <span class="max-w-105 truncate text-error" title={app.sync.error}>
          {app.sync.error}
        </span>
      </Show>

      <Show when={currentFeed()?.lastFetchedAt}>
        {(fetched) => (
          <span class="shrink-0 tabular-nums">上次刷新 {formatRelative(fetched())}</span>
        )}
      </Show>
    </footer>
  );
}
