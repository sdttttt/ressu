import { createSlice } from "@reduxjs/toolkit";
import { RessuStore } from "./typing"

const initialState = {
    content: []
};

const messageSlice = createSlice({
    name: "message",
    initialState,
    reducers: {}
});

export const selectMessage = (store: RessuStore) => store.message;
export const selectMessageLength = (store: RessuStore) => store.message.content.length;

export default messageSlice.reducer;
