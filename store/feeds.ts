import { fetchRSSText } from "@/utils/http";
import { getFeedMeta } from "wasm"
import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, Feeds } from "./typing";
import isURL from "validator/es/lib/isURL";
import { runWASM } from "@/utils/wasm";
import { toaster } from "evergreen-ui";

const initialState: Feeds = {
	channels: []
};

/**
 * async add new rss channel.
 */
export const addRSSChannelAsync = createAsyncThunk(
	"channels/add",
	async (url: string) => {
		if (isURL(url)) {
			const rssText = await fetchRSSText(url);
			const metaInfo = await runWASM(() => getFeedMeta(rssText));
			if (metaInfo.isSpecification()) {
				console.log(metaInfo.json());
				return metaInfo.json();
			} else {
				toaster.danger("该订阅源不是RSS2.0标准, 尚不支持其他标准的RSS.");
			}
		}
	}
);

const feedsSlice = createSlice({
	name: "channels",
	initialState,
	reducers: {
		initialize: (state, { payload }: PayloadAction<Feeds>) => {
			state.channels = payload.channels;
		},

		remove: (state, url: PayloadAction<string>) => {
			const { channels } = state;
			state.channels = channels.filter(rss => rss.url != url.payload);
		}
	},

	extraReducers: builder => {
		builder.addCase(addRSSChannelAsync.fulfilled, (state, actions) => {
			console.log(actions.payload, state);
		});
	}
});

export const selectFeeds = (state: RessuStore) => state.feeds;
export const selectChannels = (state: RessuStore) => state.feeds.channels;
export const selectChannelLength = (state: RessuStore) =>
	state.feeds.channels.length;

export const { remove } = feedsSlice.actions;

export default feedsSlice.reducer;
