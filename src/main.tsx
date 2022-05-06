import { getFeedMeta } from "wasm";

import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";
import { ToastContainer } from "react-toastify";
import { App } from "./app";
import store from "@store/index";
import "./index.scss";

import "react-toastify/dist/ReactToastify.css";

import "./init";

render(
	<React.StrictMode>
		<Provider store={store}>
			<App />
		</Provider>

		<ToastContainer></ToastContainer>
	</React.StrictMode>,
	document.getElementById("root")
);
