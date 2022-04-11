import { hello } from "wasm";

import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";

import { App } from "./app";
import store from "@store/index";
import "./index.scss";
import { runWASM } from "./utils/wasm";
import { fetchRSSText } from "@/utils/http";

runWASM(() => {
	hello();
	fetchRSSText("https://www.yystv.cn/rss/feed");
});

render(
	<React.StrictMode>
		<Provider store={store} >
				<App />
		</Provider>
	</React.StrictMode>,
	document.getElementById("root")
);
