import initWasm, { hello, http_get as httpGet } from "wasm";

import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";

import { App } from "./app";
import store from "@store/index";
import "./index.scss";


+(async () => {

	await initWasm();
	hello();
	httpGet("https://hyper.rs");
	
})();

render(
	<React.StrictMode>
		<Provider store={store} >
				<App />
		</Provider>
	</React.StrictMode>,
	document.getElementById("root")
);
