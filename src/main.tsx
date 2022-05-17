import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";
import { Toaster } from "react-hot-toast";
import { App } from "./app";
import store from "@store/index";
import "./index.scss";

import "./init";

render(
	<React.StrictMode>
		<Provider store={store}>
			<App />
		</Provider>
		<Toaster position="bottom-right"></Toaster>
	</React.StrictMode>,
	document.getElementById("root")
);
