import { Loader2 } from "lucide-solid";
import { Show } from "solid-js";

import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Separator } from "~/components/ui/separator";
import { Tooltip, TooltipContent, TooltipTrigger } from "~/components/ui/tooltip";
import { formatRelative } from "~/lib/format";
import { app, currentFeed } from "~/stores/app";

export default function StatusBar() {
  return (
    <footer class="flex h-7 shrink-0 items-center gap-2 border-t bg-background px-3 text-[0.6875rem] text-muted-foreground">
      <Show
        when={app.sync.active}
        fallback={
          <div class="flex items-center gap-1.5">
            <span class="tabular-nums">{app.feeds.length} 个订阅源</span>
            <Separator orientation="vertical" class="h-3" />
            <Badge
              variant="secondary"
              class="h-4 rounded-full px-1.5 py-0 text-[0.625rem] font-medium tabular-nums"
            >
              {app.totalUnread} 未读
            </Badge>
          </div>
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
        <Tooltip placement="top">
          <TooltipTrigger
            as={Button}
            variant="ghost"
            size="sm"
            class="h-5 px-1.5 text-[0.625rem] font-normal text-error"
          >
            刷新失败
          </TooltipTrigger>
          <TooltipContent class="max-w-80">{app.sync.error}</TooltipContent>
        </Tooltip>
      </Show>

      <Show when={currentFeed()?.lastFetchedAt}>
        {(fetched) => (
          <>
            <Separator orientation="vertical" class="h-3" />
            <span class="shrink-0 tabular-nums">上次刷新 {formatRelative(fetched())}</span>
          </>
        )}
      </Show>
    </footer>
  );
}
