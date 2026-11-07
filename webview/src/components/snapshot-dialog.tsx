import { ClipboardCopy, Download } from "lucide-solid";
import { Show } from "solid-js";

import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { app, closeSnapshot, copySnapshot, saveSnapshot } from "~/stores/app";

/** Shows the finished snapshot before the user chooses a destination. */
export default function SnapshotDialog() {
  return (
    <Dialog
      open={app.snapshot !== null}
      onOpenChange={(open) => {
        if (!open && app.snapshotAction === null) closeSnapshot();
      }}
    >
      <DialogContent class="max-w-4xl">
        <DialogHeader>
          <DialogTitle>文章快照</DialogTitle>
          <DialogDescription>快照已生成，可以先预览再保存或复制。</DialogDescription>
        </DialogHeader>

        <Show when={app.snapshot}>
          {(snapshot) => (
            <div class="max-h-[52vh] min-h-0 overflow-auto rounded-md border bg-muted/40 p-3">
              <img
                alt={`${snapshot().post.title} 快照预览`}
                class="mx-auto block w-full"
                src={`data:image/png;base64,${snapshot().png}`}
              />
            </div>
          )}
        </Show>

        <DialogFooter class="gap-2">
          <Button
            disabled={app.snapshotAction !== null}
            type="button"
            variant="ghost"
            onClick={closeSnapshot}
          >
            关闭
          </Button>
          <Button
            disabled={app.snapshotAction !== null}
            type="button"
            variant="outline"
            onClick={() => void copySnapshot()}
          >
            <ClipboardCopy class="size-4" />
            {app.snapshotAction === "copy" ? "复制中…" : "复制到剪贴板"}
          </Button>
          <Button
            disabled={app.snapshotAction !== null}
            type="button"
            onClick={() => void saveSnapshot()}
          >
            <Download class="size-4" />
            {app.snapshotAction === "save" ? "保存中…" : "保存为图片…"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
