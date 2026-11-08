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
import { Label } from "~/components/ui/label";
import {
  NumberField,
  NumberFieldDecrementTrigger,
  NumberFieldGroup,
  NumberFieldIncrementTrigger,
  NumberFieldInput,
  NumberFieldLabel,
} from "~/components/ui/number-field";
import { RadioGroup, RadioGroupItem, RadioGroupItemLabel } from "~/components/ui/radio-group";
import { Separator } from "~/components/ui/separator";
import { Slider, SliderFill, SliderThumb, SliderTrack } from "~/components/ui/slider";
import { Switch, SwitchControl, SwitchLabel, SwitchThumb } from "~/components/ui/switch";
import { app, closeDialog, exportOPML, openDataDir, openDialog, updatePrefs } from "~/stores/app";

const themes = [
  { value: "system", label: "跟随系统" },
  { value: "light", label: "浅色" },
  { value: "dark", label: "深色" },
  { value: "sepia", label: "护眼" },
];

const fontSizes = [
  { value: "small", label: "小" },
  { value: "medium", label: "中" },
  { value: "large", label: "大" },
];

const fontSizeIndex = () =>
  Math.max(
    0,
    fontSizes.findIndex((size) => size.value === app.prefs.fontSize),
  );

function changeFontSize([index]: number[]): void {
  const size = fontSizes[index];
  if (size && size.value !== app.prefs.fontSize) void updatePrefs({ fontSize: size.value });
}

const linkTargets = [
  { value: "browser", label: "系统浏览器" },
  { value: "window", label: "应用内新窗口" },
];

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
          <h3 class="text-[0.6875rem] font-medium text-muted-foreground">外观</h3>
          <div class="space-y-2">
            <div class="flex items-center justify-between gap-4">
              <Label class="text-sm font-normal">界面字体大小</Label>
              <span class="text-xs text-muted-foreground">{fontSizes[fontSizeIndex()]?.label}</span>
            </div>
            <Slider
              value={[fontSizeIndex()]}
              minValue={0}
              maxValue={fontSizes.length - 1}
              step={1}
              getValueLabel={({ values }) => fontSizes[values[0]]?.label ?? ""}
              onChange={changeFontSize}
              class="gap-2"
            >
              <SliderTrack class="h-1.5 w-full">
                <SliderFill />
              </SliderTrack>
              <SliderThumb class="top-[-7px]" />
              <div class="flex w-full justify-between text-xs text-muted-foreground">
                <For each={fontSizes}>{(size) => <span>{size.label}</span>}</For>
              </div>
            </Slider>
          </div>

          <Separator />

          <div class="space-y-2">
            <Label class="text-sm font-normal">主题</Label>
            <RadioGroup
              value={app.prefs.theme}
              onChange={(theme) => theme && void updatePrefs({ theme })}
              class="gap-2"
            >
              <For each={themes}>
                {(theme) => (
                  <RadioGroupItem value={theme.value} class="gap-2">
                    <RadioGroupItemLabel class="font-normal">{theme.label}</RadioGroupItemLabel>
                  </RadioGroupItem>
                )}
              </For>
            </RadioGroup>
          </div>
        </section>

        <Separator />

        <section class="space-y-3">
          <h3 class="text-[0.6875rem] font-medium text-muted-foreground">阅读</h3>

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

          <NumberField
            value={app.prefs.postsPerFeed}
            minValue={50}
            maxValue={2000}
            step={50}
            onRawValueChange={(value) => {
              if (value !== app.prefs.postsPerFeed) void updatePrefs({ postsPerFeed: value });
            }}
          >
            <div class="flex items-center justify-between gap-4">
              <NumberFieldLabel class="text-sm font-normal">
                每个订阅源保留的文章数
              </NumberFieldLabel>
              <NumberFieldGroup class="w-28">
                <NumberFieldInput class="h-8 pr-6 text-right text-xs tabular-nums" />
                <NumberFieldIncrementTrigger />
                <NumberFieldDecrementTrigger />
              </NumberFieldGroup>
            </div>
          </NumberField>
        </section>

        <Separator />

        <section class="space-y-2">
          <h3 class="text-[0.6875rem] font-medium text-muted-foreground">链接</h3>
          <RadioGroup
            value={app.prefs.openLinks}
            onChange={(openLinks) => openLinks && void updatePrefs({ openLinks })}
            class="gap-2"
          >
            <For each={linkTargets}>
              {(target) => (
                <RadioGroupItem value={target.value} class="gap-2">
                  <RadioGroupItemLabel class="font-normal">{target.label}</RadioGroupItemLabel>
                </RadioGroupItem>
              )}
            </For>
          </RadioGroup>
        </section>

        <Separator />

        <section class="space-y-2">
          <h3 class="text-[0.6875rem] font-medium text-muted-foreground">数据</h3>
          <div class="flex items-center justify-between gap-4">
            <Label class="text-sm font-normal">订阅源备份</Label>
            <Button
              variant="outline"
              size="sm"
              class="h-8 shrink-0 text-xs"
              onClick={() => void exportOPML()}
            >
              导出 OPML
            </Button>
          </div>
          <Label class="text-xs font-normal text-muted-foreground">数据目录</Label>
          <div class="flex items-start gap-2">
            <code class="min-w-0 flex-1 rounded-md bg-muted px-2 py-1.5 text-[0.6875rem] break-all">
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
