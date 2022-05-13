import { configureStore } from "@reduxjs/toolkit";
import feeds from "./feeds";
import uiState from "./ui-state";
import settings from "./settings";

export default configureStore({
	reducer: {
		settings,
		feeds,
		"ui-state": uiState
	}
});
