import { configureStore } from "@reduxjs/toolkit";
import wasm from "./wasm";
import feeds from "./feeds";
import message from "./message";
import settings from "./settings";

export default configureStore({
    reducer: {
        wasm, settings, feeds, message
    }
});
