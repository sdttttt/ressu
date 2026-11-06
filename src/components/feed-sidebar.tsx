import {
  CheckCheck,
  Loader2,
  MoreHorizontal,
  Plus,
  RefreshCw,
  Rss,
  Search,
  Settings as SettingsIcon,
  Trash2,
} from "lucide-solid";
import { For, Show, createSignal } from "solid-js";

import { EmptyState } from "~/components/empty-state";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import { Input } from "~/components/ui/input";
import { cn } from "~/lib/utils";
import type { Feed } from "~/mygo";
import {
  app,
  isSyncingFeed,
  markAllRead,
  openDialog,
  refreshAll,
  refreshFeed,
  removeFeed,
  selectFeed,
  setFeedFilter,
  visibleFeeds,
} from "~/stores/app";

export default function FeedSidebar() {
  return (
    <aside class="flex h-full min-h-0 flex-col border-r bg-background">
      <div class="flex items-center gap-1 px-2 py-2">
        <div class="relative min-w-0 flex-1">
          <Search class="pointer-events-none absolute top-1/2 left-2 size-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            class="h-8 pl-7 text-xs"
            placeholder="搜索订阅源"
            value={app.feedFilter}
            onInput={(event) => setFeedFilter(event.currentTarget.value)}
          />
        </div>
        <Button
          variant="ghost"
          size="icon"
          class="size-8 shrink-0"
          title="添加订阅源"
          onClick={() => openDialog("addFeed")}
        >
          <Plus class="size-4" />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="size-8 shrink-0"
          title="刷新全部"
          disabled={app.sync.active}
          onClick={() => void refreshAll()}
        >
          <RefreshCw class={cn("size-4", app.sync.active && "animate-spin")} />
        </Button>
        <Button
          variant="ghost"
          size="icon"
          class="size-8 shrink-0"
          title="设置"
          onClick={() => openDialog("settings")}
        >
          <SettingsIcon class="size-4" />
        </Button>
      </div>

      <div class="flex items-center justify-between px-3 pb-1 text-[11px] font-medium text-muted-foreground">
        <span>订阅源</span>
        <span>{app.feeds.length}</span>
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
          <For each={visibleFeeds()}>{(feed) => <FeedRow feed={feed} />}</For>
        </Show>
      </div>
    </aside>
  );
}

function FeedRow(props: { feed: Feed }) {
  const selected = () => app.current.feedId === props.feed.id;
  const open = () => void selectFeed(props.feed.id);

  return (
    <div
      class={cn(
        "group relative flex items-center gap-1 rounded-md pr-1 pl-1 text-[13px] transition-colors hover:bg-accent/70",
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
        onClick={open}
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
          <span class="shrink-0 rounded-full bg-primary/12 px-1.5 py-px text-[10px] font-semibold text-primary tabular-nums">
            {props.feed.unread > 99 ? "99+" : props.feed.unread}
          </span>
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
          <DropdownMenuItem onSelect={() => void removeFeed(props.feed.id)}>
            <Trash2 class="size-3.5" />
            取消订阅
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
    </div>
  );
}

function FeedIcon(props: { feed: Feed }) {
  const [failed, setFailed] = createSignal(false);

  const source = () => {
    if (props.feed.iconUrl) return props.feed.iconUrl;
    if (!props.feed.siteUrl) return "";
    try {
      return new URL("/favicon.ico", props.feed.siteUrl).toString();
    } catch {
      return "";
    }
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
