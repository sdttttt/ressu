import { getFeedMeta } from "wasm";
import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, RSSChannel } from "./typing"
import isURL from "validator/es/lib/isURL";
import { runWASM } from "@/utils/wasm";
import { toaster } from "evergreen-ui";

const initialState: RSSChannel[] = [];

/**
 * async add new rss channel.
 */
export const addRSSChannelAsync = createAsyncThunk('channels/add', async (url: string) => {
    if (isURL(url)) {
        const metaInfo = await runWASM(() => getFeedMeta(url));
        if (metaInfo.isSpecification()) {
            console.log(metaInfo.json());
            return metaInfo;
        } else {
            toaster.danger("该URL不是规范的RSS订阅源。")
        }
    }
})

const channelSlice = createSlice({
    name: "channels",
    initialState,
    reducers: {
        remove: (channels, url: PayloadAction<string>) => {
            return channels.filter(rss => rss.url != url.payload);
        },
    },

    extraReducers: (builder) => {
        builder.addCase(addRSSChannelAsync.fulfilled, (channels, actions) => {
            console.log(actions.payload);
        })
    }
});

export const selectChannels = (state: RessuStore) => state.channels;
export const selectChannelsCount = (state: RessuStore) => state.channels.length;

export const { remove } = channelSlice.actions;

export default channelSlice.reducer;