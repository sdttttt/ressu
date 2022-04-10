import { hello } from "wasm";

import { fetch, ResponseType } from "@tauri-apps/api/http"

import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";

import { App } from "./app";
import store from "@store/index";
import "./index.scss";
import { runWASM } from "./utils/wasm";

runWASM(() => {
	hello();
	fetch("https://www.yystv.cn/rss/feed", { responseType: ResponseType.Text, method: "GET" })
	.then(res => console.log(res.data));
});

render(
	<React.StrictMode>
		<Provider store={store} >
				<App />
		</Provider>
	</React.StrictMode>,
	document.getElementById("root")
);
