import { Loader2 } from "lucide-solid";
import { Show, createSignal } from "solid-js";

import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import { Input } from "~/components/ui/input";
import { addFeed, app, closeDialog, openDialog } from "~/stores/app";

export default function AddFeedDialog() {
  const [feedURL, setFeedURL] = createSignal("");
  const [error, setError] = createSignal("");
  const [pending, setPending] = createSignal(false);

  const close = () => {
    closeDialog("addFeed");
    setFeedURL("");
    setError("");
    setPending(false);
  };

  const submit = async (event: SubmitEvent) => {
    event.preventDefault();
    const url = feedURL().trim();

    if (url === "") {
      setError("请输入订阅地址");
      return;
    }
    if (!/^https?:\/\//i.test(url)) {
      setError("地址需要以 http:// 或 https:// 开头");
      return;
    }
    if (pending()) return;

    setError("");
    setPending(true);
    const added = await addFeed(url);
    setPending(false);
    if (added) setFeedURL("");
  };

  return (
    <Dialog
      open={app.dialog.addFeed}
      onOpenChange={(open) => {
        if (open) openDialog("addFeed");
        else close();
      }}
    >
      <DialogContent class="max-w-md">
        <DialogHeader>
          <DialogTitle>添加订阅源</DialogTitle>
          <DialogDescription>
            输入 RSS、Atom 或 JSON Feed 的地址，Ressu 会先抓取一次确认能读。
          </DialogDescription>
        </DialogHeader>

        <form class="space-y-3" onSubmit={submit}>
          <Input
            autofocus
            value={feedURL()}
            placeholder="https://example.com/feed.xml"
            onInput={(event) => {
              setFeedURL(event.currentTarget.value);
              setError("");
            }}
          />

          <Show when={error()}>
            <p class="text-xs text-error">{error()}</p>
          </Show>

          <DialogFooter class="gap-2">
            <Button type="button" variant="ghost" onClick={close}>
              取消
            </Button>
            <Button type="submit" disabled={pending()}>
              <Show when={pending()} fallback="添加">
                <Loader2 class="size-3.5 animate-spin" />
                添加中…
              </Show>
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
