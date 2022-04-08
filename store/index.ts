import { configureStore } from "@reduxjs/toolkit";
import wasm from "./wasm";
import settings from "./settings";
import channels from "./channels";

export default configureStore({
    reducer: {
        wasm, settings, channels
    }
});