import { configureStore } from "@reduxjs/toolkit";
import feeds from "./feeds";
import uiState from "./ui-state";
import settings from "./settings";

const store =  configureStore({
	reducer: {
		settings,
		feeds,
		"ui-state": uiState
	}
});

export default store;
