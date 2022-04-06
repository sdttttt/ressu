import { configureStore } from "@reduxjs/toolkit";
import wasm from "./wasm";
import channels from "./channels";
import settings from "./settings";

const store = configureStore({
    reducer: {
        wasm, channels, settings,
    }
});


export default store;