import { createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, RSSChannel } from "./typing"

const initialState: RSSChannel[] = []; 

const channelSlice = createSlice({
    name: "channels",
    initialState,
    reducers: {
        remove: (channels, url: PayloadAction<string>) => {
            return channels.filter(rss => rss.url != url.payload);
        }
    },
});

export const selectChannels = (state: RessuStore) => state.channels;

export const { remove } = channelSlice.actions;

export default channelSlice.reducer;