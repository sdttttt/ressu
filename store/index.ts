import { configureStore } from "@reduxjs/toolkit";
import feeds from "./feeds";
import message from "./message";
import settings from "./settings";

export default configureStore({
	reducer: {
		settings,
		feeds,
		message
	}
});
