import type { Component, ComponentProps } from "solid-js";
import { splitProps } from "solid-js";

import { cn } from "~/lib/utils";

type InputProps = ComponentProps<"input"> & { class?: string | undefined };

const Input: Component<InputProps> = (props) => {
  const [local, others] = splitProps(props, ["class"]);
  return (
    <input
      class={cn(
        "flex h-9 w-full min-w-0 rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs transition-colors placeholder:text-muted-foreground focus-visible:ring-1 focus-visible:ring-ring focus-visible:outline-none disabled:cursor-not-allowed disabled:opacity-50",
        local.class,
      )}
      {...others}
    />
  );
};

export { Input };
export type { InputProps };
