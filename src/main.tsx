import { getFeedMeta } from "wasm";

import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";

import { App } from "./app";
import store from "@store/index";
import "./index.scss";
import { runWASM } from "./utils/wasm";
import { fetchRSSText } from "@/utils/http";

import { ToastContainer } from "react-toastify";
import "react-toastify/dist/ReactToastify.css";

runWASM(async () => {
	const text = await fetchRSSText("https://www.yystv.cn/rss/feed");
	const rss_channel = await getFeedMeta(text);
	console.log(rss_channel);
});

render(
	<React.StrictMode>
		<Provider store={store}>
			<App />
		</Provider>

		<ToastContainer></ToastContainer>
	</React.StrictMode>,
	document.getElementById("root")
);
