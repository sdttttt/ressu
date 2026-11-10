import {
  CheckCheck,
  ChevronLeft,
  ChevronRight,
  Loader2,
  MoreHorizontal,
  Plus,
  RefreshCw,
  Rss,
  Search,
  Settings as SettingsIcon,
  Trash2,
} from "lucide-solid";
import { For, Show, createEffect, createSignal } from "solid-js";

import { EmptyState } from "~/components/empty-state";
import { IconButton } from "~/components/icon-button";
import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogTitle,
} from "~/components/ui/alert-dialog";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { Popover, PopoverContent, PopoverTrigger } from "~/components/ui/popover";
import { TextField, TextFieldInput } from "~/components/ui/text-field";
import { cn } from "~/lib/utils";
import type { Feed } from "~/mygo";
import {
  app,
  collapseSidebar,
  expandSidebar,
  ensureFeedIcon,
  isSyncingFeed,
  markAllRead,
  openDialog,
  openSettings,
  refreshAll,
  refreshFeed,
  removeFeed,
  selectFeed,
  setFeedFilter,
  visibleFeeds,
} from "~/stores/app";

export default function FeedSidebar() {
  return (
    <aside class="flex h-full min-h-0 w-full flex-col bg-background">
      <div class="flex items-center gap-1 px-3 pt-2 pb-1">
        <div class="flex min-w-0 flex-1 items-center gap-1.5">
          <span class="truncate text-[0.8125rem] font-semibold">订阅源</span>
          <Badge
            variant="outline"
            class="h-4 shrink-0 rounded-full px-1.5 py-0 text-[0.625rem] tabular-nums"
          >
            {app.feeds.length}
          </Badge>
        </div>

        <IconButton
          label={app.sidebar.collapsed ? "恢复三栏布局" : "浮动显示订阅源"}
          class="size-7 shrink-0"
          onClick={() => (app.sidebar.collapsed ? expandSidebar() : collapseSidebar())}
        >
          {app.sidebar.collapsed ? <ChevronRight class="size-4" /> : <ChevronLeft class="size-4" />}
        </IconButton>

        <Popover placement="bottom-end">
          <PopoverTrigger
            as={Button}
            variant="ghost"
            size="icon"
            class="size-7 shrink-0"
            aria-label="搜索订阅源"
          >
            <Search class="size-4" />
          </PopoverTrigger>
          <PopoverContent class="w-64 p-2">
            <TextField value={app.feedFilter} onChange={setFeedFilter}>
              <TextFieldInput autofocus placeholder="按名称或地址过滤" class="h-8 text-xs" />
            </TextField>
          </PopoverContent>
        </Popover>

        <IconButton label="添加订阅源" class="size-7" onClick={() => openDialog("addFeed")}>
          <Plus class="size-4" />
        </IconButton>

        <IconButton
          label="刷新全部"
          class="size-7"
          disabled={app.sync.active}
          onClick={() => void refreshAll()}
        >
          <RefreshCw class={cn("size-4", app.sync.active && "animate-spin")} />
        </IconButton>

        <IconButton label="设置" class="size-7" onClick={() => void openSettings()}>
          <SettingsIcon class="size-4" />
        </IconButton>
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto px-2 pb-3">
        <Show
          when={visibleFeeds().length > 0}
          fallback={
            <Show
              when={app.feeds.length === 0}
              fallback={<EmptyState title="没有匹配的订阅源" hint="换一个关键字试试。" />}
            >
              <EmptyState
                icon={<Rss class="size-10" />}
                title="还没有订阅源"
                hint="点击上方的 + 添加一个 RSS、Atom 或 JSON Feed 地址。"
              />
            </Show>
          }
        >
          <div class="space-y-0.5">
            <For each={visibleFeeds()}>{(feed) => <FeedRow feed={feed} />}</For>
          </div>
        </Show>
      </div>
    </aside>
  );
}

function FeedRow(props: { feed: Feed }) {
  const selected = () => app.current.feedId === props.feed.id;
  const toggle = () => void selectFeed(props.feed.id);
  const open = () => {
    if (!selected()) void selectFeed(props.feed.id);
  };
  const [confirming, setConfirming] = createSignal(false);

  return (
    <div
      class={cn(
        "group relative flex items-center gap-1 rounded-md pr-1 pl-1 text-[0.8125rem] transition-colors hover:bg-accent/70",
        selected() && "bg-accent",
      )}
    >
      <Show when={selected()}>
        <span class="absolute top-1/2 left-0.5 h-3.5 w-0.5 -translate-y-1/2 rounded-full bg-primary" />
      </Show>

      <button
        type="button"
        class="flex min-w-0 flex-1 items-center gap-2 py-1.5 pl-1.5 text-left"
        title={props.feed.title}
        onClick={toggle}
      >
        <FeedIcon feed={props.feed} />

        <span class={cn("min-w-0 flex-1 truncate", selected() && "font-medium")}>
          {props.feed.title}
        </span>

        <Show when={isSyncingFeed(props.feed.id)}>
          <Loader2 class="size-3.5 shrink-0 animate-spin text-muted-foreground" />
        </Show>

        <Show when={props.feed.lastError !== ""}>
          <span class="size-1.5 shrink-0 rounded-full bg-error" title={props.feed.lastError} />
        </Show>

        <Show when={props.feed.unread > 0}>
          <Badge
            variant="secondary"
            class="h-4 shrink-0 rounded-full px-1.5 py-0 text-[0.625rem] font-semibold tabular-nums"
          >
            {props.feed.unread > 99 ? "99+" : props.feed.unread}
          </Badge>
        </Show>
      </button>

      <DropdownMenu>
        <DropdownMenuTrigger
          as={Button}
          variant="ghost"
          size="icon"
          class="size-6 shrink-0 opacity-0 group-hover:opacity-100 data-[expanded]:opacity-100"
          aria-label={`${props.feed.title} 的更多操作`}
        >
          <MoreHorizontal class="size-3.5" />
        </DropdownMenuTrigger>
        <DropdownMenuContent class="w-44">
          <DropdownMenuItem onSelect={open}>打开</DropdownMenuItem>
          <DropdownMenuItem onSelect={() => void refreshFeed(props.feed.id)}>
            <RefreshCw class="size-3.5" />
            刷新
          </DropdownMenuItem>
          <DropdownMenuItem onSelect={() => void markAllRead(props.feed.id)}>
            <CheckCheck class="size-3.5" />
            标记全部已读
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem onSelect={() => setConfirming(true)}>
            <Trash2 class="size-3.5" />
            取消订阅
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>

      <AlertDialog open={confirming()} onOpenChange={setConfirming}>
        <AlertDialogContent class="max-w-md">
          <AlertDialogTitle>取消订阅「{props.feed.title}」？</AlertDialogTitle>
          <AlertDialogDescription>
            这个订阅源已保存的文章会一并删除，操作无法撤销。
          </AlertDialogDescription>
          <div class="flex justify-end gap-2">
            <Button variant="ghost" onClick={() => setConfirming(false)}>
              取消
            </Button>
            <Button
              variant="destructive"
              onClick={() => {
                setConfirming(false);
                void removeFeed(props.feed.id);
              }}
            >
              取消订阅
            </Button>
          </div>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

function FeedIcon(props: { feed: Feed }) {
  const [failed, setFailed] = createSignal(false);

  createEffect(() => void ensureFeedIcon(props.feed.id));

  const source = () => {
    const icon = app.icons[props.feed.id];
    return icon?.missing || !icon?.src ? "" : icon.src;
  };

  return (
    <Show
      when={source() !== "" && !failed()}
      fallback={<Rss class="size-4 shrink-0 text-muted-foreground" />}
    >
      <img
        src={source()}
        alt=""
        class="size-4 shrink-0 rounded-sm object-cover"
        loading="lazy"
        onError={() => setFailed(true)}
      />
    </Show>
  );
}
