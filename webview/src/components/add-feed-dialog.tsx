import { Loader2 } from "lucide-solid";
import { Show, createSignal } from "solid-js";

import { Alert, AlertDescription } from "~/components/ui/alert";
import { Button } from "~/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/components/ui/dialog";
import {
  TextField,
  TextFieldDescription,
  TextFieldInput,
  TextFieldLabel,
} from "~/components/ui/text-field";
import { addFeed, app, closeDialog, importOPML, openDialog, testFeed } from "~/stores/app";

export default function AddFeedDialog() {
  const [feedURL, setFeedURL] = createSignal("");
  const [feedName, setFeedName] = createSignal("");
  const [error, setError] = createSignal("");
  const [testedURL, setTestedURL] = createSignal("");
  const [pending, setPending] = createSignal(false);
  const [testing, setTesting] = createSignal(false);

  const close = () => {
    closeDialog("addFeed");
    setFeedURL("");
    setFeedName("");
    setError("");
    setTestedURL("");
    setPending(false);
    setTesting(false);
  };

  const validURL = () => {
    const url = feedURL().trim();

    if (url === "") {
      setError("请输入订阅地址");
      return null;
    }
    if (!/^https?:\/\//i.test(url)) {
      setError("地址需要以 http:// 或 https:// 开头");
      return null;
    }
    return url;
  };

  const test = async () => {
    const url = validURL();
    if (url === null || testing() || pending()) return;

    setError("");
    setTesting(true);
    const result = await testFeed(url);
    setTesting(false);
    if (url !== feedURL().trim()) return;

    if (result.error !== "") {
      setTestedURL("");
      setError(result.error);
      return;
    }

    setTestedURL(url);
    if (feedName().trim() === "") {
      setFeedName(result.feed?.title ?? url);
    }
  };

  const submit = async (event: SubmitEvent) => {
    event.preventDefault();
    const url = validURL();
    if (url === null) return;
    if (pending()) return;

    setError("");
    setPending(true);
    const added = await addFeed(feedName(), url);
    setPending(false);
    if (added) close();
  };

  const canAdd = () => {
    const url = feedURL().trim();
    return url !== "" && testedURL() === url;
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
          <TextField
            value={feedURL()}
            onChange={(value) => {
              setFeedURL(value);
              setError("");
              setTestedURL("");
            }}
            class="gap-1.5"
          >
            <TextFieldLabel class="text-xs font-normal text-muted-foreground">
              订阅地址
            </TextFieldLabel>
            <TextFieldInput type="url" autofocus placeholder="https://example.com/feed.xml" />
            <TextFieldDescription class="text-[0.6875rem]">
              支持 RSS 0.9x / 1.0 / 2.0、Atom 与 JSON Feed。
            </TextFieldDescription>
          </TextField>

          <TextField value={feedName()} onChange={(value) => setFeedName(value)} class="gap-1.5">
            <TextFieldLabel class="text-xs font-normal text-muted-foreground">名称</TextFieldLabel>
            <TextFieldInput type="text" placeholder="留空则使用订阅源名称" />
          </TextField>

          <Show
            when={error()}
            fallback={
              <Show when={testedURL()}>
                <output class="text-xs text-success">测试通过，可以添加。</output>
              </Show>
            }
          >
            <Alert variant="destructive" class="px-3 py-2">
              <AlertDescription class="text-xs">{error()}</AlertDescription>
            </Alert>
          </Show>

          <DialogFooter class="gap-2">
            <Button
              type="button"
              variant="ghost"
              class="mr-auto text-xs text-muted-foreground"
              onClick={() => void importOPML()}
            >
              从 OPML 导入
            </Button>
            <Button
              type="button"
              variant="outline"
              disabled={testing() || pending()}
              onClick={() => void test()}
            >
              <Show when={testing()} fallback="测试">
                <Loader2 class="size-3.5 animate-spin" />
                测试中…
              </Show>
            </Button>
            <Button type="submit" disabled={pending() || !canAdd()}>
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
