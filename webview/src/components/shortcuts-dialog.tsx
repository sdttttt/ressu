import { For, Show } from "solid-js";

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { commands, formatBinding } from "~/lib/commands";
import { app, closeDialog, openDialog } from "~/stores/app";

/** The groups the help dialog shows, in the commands' declaration order. */
const groups = ["应用", "文章列表", "阅读", "订阅源"] as const;

export default function ShortcutsDialog() {
  return (
    <Dialog
      open={app.dialog.help}
      onOpenChange={(open) => {
        if (open) openDialog("help");
        else closeDialog("help");
      }}
    >
      <DialogContent class="max-w-lg">
        <DialogHeader>
          <DialogTitle>键盘快捷键</DialogTitle>
          <DialogDescription>
            单字母快捷键在输入框和对话框打开时不起作用；macOS 上 ⌘ 开头的组合由系统菜单接管。
          </DialogDescription>
        </DialogHeader>

        <div class="min-h-0 space-y-4">
          <For each={[...groups]}>
            {(group) => (
              <Show when={commands.some((command) => command.group === group)}>
                <section class="space-y-1">
                  <h3 class="text-[0.6875rem] font-semibold tracking-wide text-muted-foreground">
                    {group}
                  </h3>
                  <ul class="divide-y divide-border/60">
                    <For each={commands.filter((command) => command.group === group)}>
                      {(command) => (
                        <li class="flex items-center justify-between gap-3 py-1.5 text-sm">
                          <span class="min-w-0 truncate">{command.label}</span>
                          <kbd class="shrink-0 rounded border bg-muted px-1.5 py-0.5 font-mono text-[0.6875rem] text-foreground">
                            {formatBinding(command.binding)}
                          </kbd>
                        </li>
                      )}
                    </For>
                  </ul>
                </section>
              </Show>
            )}
          </For>
        </div>
      </DialogContent>
    </Dialog>
  );
}
