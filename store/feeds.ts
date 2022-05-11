import { emit } from "@tauri-apps/api/event";
import { feedsDataLocalGet, feedsDataLocalSync } from "@database/feeds";
import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, Feeds, RSSChannel } from "./typing";
import isURL from "validator/es/lib/isURL";
import { parseRSSFromURL } from "@/utils/wasm";
import { toast } from "react-toastify";
import { RessuEvent } from "@/listens";
import { concat, uniqWith } from "lodash-es";

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
	async (channels: RSSChannel[]): Promise<(RSSChannel|undefined)[]> => {
		const promiseResults: Promise<RSSChannel|undefined>[] = channels
		.map(async (ch): Promise<RSSChannel|undefined> => {
			const result = await parseRSSFromURL(ch.url);
			return result;
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
		initialize: (state: Feeds, { payload }: PayloadAction<Feeds>) => {
			state.channels = payload.channels;
		},

		remove: (state: Feeds , url: PayloadAction<string>) => {
			const { channels } = state;
			state.channels = channels.filter(rss => rss.url != url.payload);
		}
	},

	extraReducers: builder => {

		builder.addCase(addRSSChannelAsync.fulfilled, (state, { payload: channel }) => {
			console.log(channel, state);
			if (channel) {

				if (state.channels.map(t => t.url).includes(channel.url)) return;

				state.channels.push(channel);
				feedsDataLocalSync(state);
				return state;
			} else {
				toast.error("无效的RSS订阅信息。");
			}
		});


		builder.addCase(pullRSSChannelAsync.fulfilled, (state, { payload: newCh }) => {
			const { channels } = state;
			for (let x = 0; x < channels.length; x++) {
				for (let y = 0; y < newCh.length; y++) {
					if (newCh[y] === undefined) continue;
					if (newCh[y]!.url === channels[x].url) {
						channels[x].posts = concat(newCh[y]!.posts, channels[x].posts);
						channels[x].posts = uniqWith(channels[x].posts, (a, b) => a.guid === b.guid && a.title === b.title );
					}
				}
			}
			state.channels = channels;
			return state;
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
