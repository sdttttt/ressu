import type { JSX } from "solid-js";

import { Button } from "~/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "~/components/ui/tooltip";
import { cn } from "~/lib/utils";

type IconButtonProps = {
  label: string;
  disabled?: boolean;
  class?: string;
  onClick?: (event: MouseEvent) => void;
  children: JSX.Element;
};

/** An icon-only solid-ui button whose tooltip is also its accessible name. */
export function IconButton(props: IconButtonProps) {
  return (
    <Tooltip placement="bottom">
      <TooltipTrigger
        as={Button}
        variant="ghost"
        size="icon"
        aria-label={props.label}
        disabled={props.disabled}
        class={cn("size-8 shrink-0", props.class)}
        onClick={props.onClick}
      >
        {props.children}
      </TooltipTrigger>
      <TooltipContent>{props.label}</TooltipContent>
    </Tooltip>
  );
}
