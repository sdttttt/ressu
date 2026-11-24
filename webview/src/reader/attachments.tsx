import { For, Show } from "solid-js";

import { Posts } from "../mygo";
import { enclosureName, enclosureSize } from "./enclosures";
import { externalHref } from "./links";
import { notifyOpenError } from "./store";
import type { Enclosure } from "../mygo";

const downloadEnclosure = async (url: string) => {
  try {
    await Posts.downloadEnclosure(url);
  } catch (error) {
    notifyOpenError("下载附件失败", error);
  }
};

// Enclosure values are untrusted feed data: Solid escapes every text
// interpolation, and every address re-passes externalHref. Audio plays
// natively in the WebView, which connects to the address directly, the
// same as media embedded in the article body; downloading goes through
// the Go binding instead.
function EnclosureRow(props: { enclosure: Enclosure }) {
  const url = externalHref(props.enclosure.url);
  if (!url) return null;
  const size = enclosureSize(props.enclosure.length);
  const label = size ? `${enclosureName(url)}（${size}）` : enclosureName(url);
  const type = props.enclosure.type || "";
  return (
    <div class="enclosure">
      <span class="enclosure-label">{label}</span>
      {type.startsWith("audio/") && <audio controls preload="none" src={url} />}
      <button
        type="button"
        class="enclosure-download"
        onClick={() => downloadEnclosure(url)}
      >
        下载附件
      </button>
    </div>
  );
}

/** The attachment bar at the end of an article body. */
export function Attachments(props: { enclosures: Enclosure[] | null }) {
  const rows = () => (Array.isArray(props.enclosures) ? props.enclosures : []);
  return (
    <Show when={rows().some((enclosure) => externalHref(enclosure.url))}>
      <div class="enclosures">
        <For each={rows()}>{(enclosure) => <EnclosureRow enclosure={enclosure} />}</For>
      </div>
    </Show>
  );
}
