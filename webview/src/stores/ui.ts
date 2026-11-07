import { setAppState, type DialogName } from "~/stores/app-state";

export function openDialog(name: DialogName): void {
  setAppState("dialog", name, true);
}

export function closeDialog(name: DialogName): void {
  setAppState("dialog", name, false);
}

export function setFeedFilter(value: string): void {
  setAppState("feedFilter", value);
}
