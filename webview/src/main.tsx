/* @refresh reload */
import { render } from "solid-js/web";

import App from "~/app";
import { applyTheme } from "~/lib/theme";

import "~/index.css";

// Paint the appearance the system asks for before the preferences arrive.
applyTheme("system");

const root = document.getElementById("root");
if (!root) throw new Error("#root is missing from index.html");

render(() => <App />, root);
