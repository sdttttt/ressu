import { fetchRSSText } from "@/utils/http";
import { feedsDataLocalGet, feedsDataLocalSync } from "@database/feeds";
import { getFeedMeta, RSSChannel as RSSChannelWasm } from "wasm";
import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, Feeds, RSSChannel } from "./typing";
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
				const resultJson = metaInfo.json();
				metaInfo.free();
				console.log(resultJson);
				return resultJson;
			} else {
				toaster.danger("Ressu目前只支持RSS2.0.");
			}
		}
	}
);

/**
 * async fetch all rss channel.
 */
export const pullRSSChannelAsync = createAsyncThunk(
	"channels/pull",
	async (channels: RSSChannel[]): Promise<[string, RSSChannelWasm][]> => {
		const promiseResults: Promise<[string, RSSChannelWasm]>[] = channels
		.map(async (ch): Promise<[string, RSSChannelWasm]> => {
			const rssText = await fetchRSSText(ch.url);
			const metaInfo = await runWASM(() => getFeedMeta(rssText));
			return [ch.url, metaInfo];
		});

		return await Promise.all(promiseResults);
	}
);

export const initinalizeFeedsFromLocal = createAsyncThunk(
	"channels/initFromLocal",
	async (): Promise<Feeds> => {
		let feedsLocal = await feedsDataLocalGet();
		console.log(feedsLocal);
		if (feedsLocal === null) {
			console.log("init: no local data.");
			await feedsDataLocalSync(initialState);
			feedsLocal = initialState;
		}

		return feedsLocal;
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
		builder.addCase(addRSSChannelAsync.fulfilled, (state, { payload }) => {
			console.log(payload, state);
		});


		builder.addCase(pullRSSChannelAsync.fulfilled, (state, { payload }) => {
			console.log(payload);
		});


		builder.addCase(
			initinalizeFeedsFromLocal.fulfilled,
			(state, { payload }: PayloadAction<Feeds>) => {
				state.channels = payload.channels;
			}
		);
	}
});

export const selectFeeds = (state: RessuStore) => state.feeds;
export const selectChannels = (state: RessuStore) => state.feeds.channels;
export const selectChannelLength = (state: RessuStore) =>
	state.feeds.channels.length;

export const { remove } = feedsSlice.actions;

export default feedsSlice.reducer;
