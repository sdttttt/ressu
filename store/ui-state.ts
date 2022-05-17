import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { UIState, RessuStore, StatusBarType, StatusBarText } from "./typing";

const initialState: UIState = {
	messages: [],
	currentChannelIndex: undefined,
	currentPostIndex: undefined,
	statusBar: {
		type: StatusBarType.Free,
		text: StatusBarText.Free,
	},
};

const messageSlice = createSlice({
	name: "ui-state",
	initialState,
	reducers: {
		send: (state: UIState, { payload }: PayloadAction<string>) => {
			state.messages.push(payload);
		},

		selectChannel: (state: UIState, { payload }: PayloadAction<number>) => {
			state.currentPostIndex = undefined;
			state.currentChannelIndex = payload;
		},

		selectPost: (state: UIState, { payload }: PayloadAction<number>) => {
			state.currentPostIndex = payload;
		},

		statusBarBuys: (state: UIState, { payload }: PayloadAction<StatusBarText>) => {
			state.statusBar = { type: StatusBarType.Buys, text: payload };
		},

		statusBarFree: (state: UIState, { payload }: PayloadAction<StatusBarText>) => {
			const { statusBar } = state;
			if (statusBar.text === payload) {
				state.statusBar = { type: StatusBarType.Free, text: StatusBarText.Free };
			}
		}
	}
});

export const selectMessage = (store: RessuStore) => store.uiState.messages;

export const selectChannelIndex = (store: RessuStore) => store.uiState.currentChannelIndex;

export const selectPostIndex = (store: RessuStore) => store.uiState.currentPostIndex;

export const selectMessageLength = (store: RessuStore) => store.uiState.messages.length;

export const { 
	send, selectChannel, selectPost,
	statusBarBuys, statusBarFree
} = messageSlice.actions;

export default messageSlice.reducer;
