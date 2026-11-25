import { isMyGo, preferences } from "mygo-runtime";
import { render } from "@solidjs/web";

import { ReaderApp } from "./app";

import "./styles.css";

if (!isMyGo()) {
  throw new Error("mygo runtime not found");
}

const desktop = preferences();
document.documentElement.style.setProperty("--mygo-text-scale", String(desktop.textScale));
if (desktop.accent) {
  document.documentElement.style.setProperty("--mygo-accent", desktop.accent);
}

const root = document.getElementById("reader");
if (!root) {
  throw new Error("missing element #reader");
}

render(() => <ReaderApp />, root);
