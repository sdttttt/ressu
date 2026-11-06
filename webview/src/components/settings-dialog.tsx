import { Settings as SettingsIcon } from "lucide-solid";
import { For } from "solid-js";

import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { Switch, SwitchControl, SwitchLabel, SwitchThumb } from "~/components/ui/switch";
import { app, closeDialog, openDataDir, openDialog, updatePrefs } from "~/stores/app";

const themes = [
  { value: "system", label: "跟随系统" },
  { value: "light", label: "浅色" },
  { value: "dark", label: "深色" },
];

const linkTargets = [
  { value: "browser", label: "系统浏览器" },
  { value: "window", label: "应用内新窗口" },
];

function commitPostsPerFeed(raw: string): void {
  const value = Number.parseInt(raw, 10);
  if (Number.isNaN(value)) return;

  const clamped = Math.min(Math.max(value, 50), 2000);
  if (clamped !== app.prefs.postsPerFeed) void updatePrefs({ postsPerFeed: clamped });
}

export default function SettingsDialog() {
  return (
    <Dialog
      open={app.dialog.settings}
      onOpenChange={(open) => (open ? openDialog("settings") : closeDialog("settings"))}
    >
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle class="flex items-center gap-2">
            <SettingsIcon class="size-4" />
            设置
          </DialogTitle>
          <DialogDescription>改动会立即保存。</DialogDescription>
        </DialogHeader>

        <section class="space-y-2">
          <h3 class="text-[11px] font-medium text-muted-foreground">外观</h3>
          <ChoiceGroup
            options={themes}
            value={app.prefs.theme}
            onSelect={(theme) => void updatePrefs({ theme })}
          />
        </section>

        <section class="space-y-3">
          <h3 class="text-[11px] font-medium text-muted-foreground">阅读</h3>

          <Switch
            class="flex items-center justify-between gap-4"
            checked={app.prefs.markReadOnOpen}
            onChange={(checked) => void updatePrefs({ markReadOnOpen: checked })}
          >
            <SwitchLabel class="text-sm font-normal">打开文章时标记为已读</SwitchLabel>
            <SwitchControl>
              <SwitchThumb />
            </SwitchControl>
          </Switch>

          <Switch
            class="flex items-center justify-between gap-4"
            checked={app.prefs.refreshOnStart}
            onChange={(checked) => void updatePrefs({ refreshOnStart: checked })}
          >
            <SwitchLabel class="text-sm font-normal">启动时刷新全部订阅源</SwitchLabel>
            <SwitchControl>
              <SwitchThumb />
            </SwitchControl>
          </Switch>

          <div class="flex items-center justify-between gap-4">
            <span id="posts-per-feed-label" class="text-sm">
              每个订阅源保留的文章数
            </span>
            <Input
              id="posts-per-feed"
              aria-labelledby="posts-per-feed-label"
              type="number"
              min="50"
              max="2000"
              step="50"
              class="h-8 w-24 text-right text-xs tabular-nums"
              value={app.prefs.postsPerFeed}
              onChange={(event) => commitPostsPerFeed(event.currentTarget.value)}
            />
          </div>
        </section>

        <section class="space-y-2">
          <h3 class="text-[11px] font-medium text-muted-foreground">链接</h3>
          <ChoiceGroup
            options={linkTargets}
            value={app.prefs.openLinks}
            onSelect={(openLinks) => void updatePrefs({ openLinks })}
          />
        </section>

        <section class="space-y-2">
          <h3 class="text-[11px] font-medium text-muted-foreground">数据</h3>
          <div class="flex items-start gap-2">
            <code class="min-w-0 flex-1 rounded-md bg-muted px-2 py-1.5 text-[11px] break-all">
              {app.dataDir || "…"}
            </code>
            <Button
              variant="outline"
              size="sm"
              class="h-8 shrink-0 text-xs"
              onClick={() => void openDataDir()}
            >
              打开目录
            </Button>
          </div>
        </section>

        <div class="flex justify-end">
          <Button variant="ghost" onClick={() => closeDialog("settings")}>
            关闭
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}

function ChoiceGroup(props: {
  options: { value: string; label: string }[];
  value: string;
  onSelect: (value: string) => void;
}) {
  return (
    <div class="flex gap-1.5">
      <For each={props.options}>
        {(option) => (
          <Button
            variant={props.value === option.value ? "default" : "outline"}
            size="sm"
            class="h-8 flex-1 text-xs font-normal"
            onClick={() => props.onSelect(option.value)}
          >
            {option.label}
          </Button>
        )}
      </For>
    </div>
  );
}
