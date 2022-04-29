import { createSlice } from "@reduxjs/toolkit";
import { RessuStore, Settings } from "./typing";

const initialState: Settings = {};

const settingsSlice = createSlice({
	name: "settings",
	initialState,
	reducers: {}
});

export const selectSettings = (state: RessuStore) => state.settings;

export const {} = settingsSlice.actions;

export default settingsSlice.reducer;
