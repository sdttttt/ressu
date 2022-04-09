import { configureStore } from "@reduxjs/toolkit";
import wasm from "./wasm";
import channels from "./channels";
import settings from "./settings";

export default configureStore({
    reducer: {
        wasm, settings, channels
    }
});