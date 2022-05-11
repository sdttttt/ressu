
import { feedsDataLocalGet, feedsDataLocalSync } from "@database/feeds";
import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, Feeds, RSSChannel } from "./typing";
import isURL from "validator/es/lib/isURL";
import { parseRSSFromURL } from "@/utils/wasm";
import { toast } from "react-toastify";

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
			return parseRSSFromURL(url);
		} else {
			toast.error("URL格式错误");
		}
	}
);

/**
 * async fetch all rss channel.
 */
export const pullRSSChannelAsync = createAsyncThunk(
	"channels/pull",
	async (channels: RSSChannel[]): Promise<[string, RSSChannel|undefined][]> => {
		const promiseResults: Promise<[string, RSSChannel|undefined]>[] = channels
		.map(async (ch): Promise<[string, RSSChannel|undefined]> => {
			const result = await parseRSSFromURL(ch.url);
			return [ch.url, result];
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
		builder.addCase(addRSSChannelAsync.fulfilled, (state, { payload: channel }) => {
			console.log(channel, state);
			if (channel) {
				state.channels.push(channel);
			} else {
				toast.error("无效的RSS订阅信息。");
			}
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
