import { isMyGo } from "mygo-runtime";
import { render } from "solid-js/web";

import { SnapshotApp } from "./app";

import "./styles.css";

if (!isMyGo()) {
  throw new Error("mygo runtime not found");
}

const root = document.getElementById("snapshot");
if (!root) {
  throw new Error("missing element #snapshot");
}

render(() => <SnapshotApp />, root);
