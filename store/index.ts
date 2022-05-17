import { configureStore } from "@reduxjs/toolkit";
import feeds from "./feeds";
import uiState from "./ui-state";
import settings from "./settings";
import { useDispatch } from "react-redux";
import { AppDispath } from "./typing";

const store =  configureStore({
	reducer: {
		settings,
		feeds,
		uiState,
	}
});

export const useAppDispatch = () =>  useDispatch<AppDispath>();

export default store;
