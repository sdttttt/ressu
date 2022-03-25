import React from "react";
import { render } from "react-dom";
import { common } from "@mui/material/colors";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import { SnackbarProvider } from "notistack";

import { App } from "./app";
import "./index.scss";

const theme = createTheme({
	palette: {
		primary: {
			main: common.black,
		},
	},
});

render(
	<React.StrictMode>
		<ThemeProvider theme={theme}>
			<SnackbarProvider maxSnack={5}>
				<App />
			</SnackbarProvider>
		</ThemeProvider>
	</React.StrictMode>,
	document.getElementById("root")
);
