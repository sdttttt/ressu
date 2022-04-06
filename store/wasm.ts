import initWasm from "wasm";
import { createSlice, createAsyncThunk } from "@reduxjs/toolkit";
import { RessuStore } from "./typing"

const initialState = {
    ready: false,
}

export const initWASMAsync = createAsyncThunk("wasm/init", async () => {
    await initWasm();
}) 

const wasmSlice = createSlice({
    name: "wasm",
    initialState,
    reducers: {},
    extraReducers: (builder) => {
        builder.addCase(initWASMAsync.fulfilled, (wasm) => {
            console.log("wasm ready.");
            wasm.ready = true;
        })
        .addCase(initWASMAsync.pending, () => {
            console.log("wasm init...")
        })
    }
});

export const selectWASM = (state: RessuStore) => state.wasm;

export const {} = wasmSlice.actions;

export default wasmSlice.reducer;