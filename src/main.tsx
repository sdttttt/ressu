import initWasm, { hello, http_get as httpGet } from "wasm";

import React from "react";
import { Provider } from "react-redux";
import { render } from "react-dom";
import { common } from "@mui/material/colors";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import { SnackbarProvider } from "notistack";

import { App } from "./app";
import store from "@store/index";
import "./index.scss";

const theme = createTheme({
	palette: {
		primary: {
			main: common.black
		}
	}
});

+(async () => {

	await initWasm();
	hello();
	httpGet("https://hyper.rs");
	
})();

render(
	<React.StrictMode>
		<Provider store={store} >
		<ThemeProvider theme={theme}>
			<SnackbarProvider maxSnack={5}>
				<App />
			</SnackbarProvider>
		</ThemeProvider>
		</Provider>
	</React.StrictMode>,
	document.getElementById("root")
);
