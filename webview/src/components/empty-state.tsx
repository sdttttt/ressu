import type { JSX } from "solid-js";
import { Show } from "solid-js";

import { cn } from "~/lib/utils";

type EmptyStateProps = {
  icon?: JSX.Element;
  title: string;
  hint?: string;
  class?: string;
};

export function EmptyState(props: EmptyStateProps) {
  return (
    <div
      class={cn(
        "flex h-full flex-col items-center justify-center gap-2 px-6 py-10 text-center",
        props.class,
      )}
    >
      <Show when={props.icon}>
        <div class="text-muted-foreground/40">{props.icon}</div>
      </Show>
      <p class="text-sm text-muted-foreground">{props.title}</p>
      <Show when={props.hint}>
        <p class="max-w-56 text-xs leading-relaxed text-muted-foreground/70">{props.hint}</p>
      </Show>
    </div>
  );
}
