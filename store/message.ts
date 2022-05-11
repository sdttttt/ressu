import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { Message, RessuStore } from "./typing";

const initialState: Message = {
	contents: []
};

const messageSlice = createSlice({
	name: "message",
	initialState,
	reducers: {
		send: (state: Message, { payload }: PayloadAction<string>) => {
			const { contents } = state;
			contents.push(payload);
		}
	}
});

export const selectMessage = (store: RessuStore) => store.message;
export const selectMessageLength = (store: RessuStore) =>
	store.message.contents.length;

export const { send } = messageSlice.actions;

export default messageSlice.reducer;
