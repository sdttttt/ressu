import { feedsDataLocalGet, feedsDataLocalSync } from "@database/feeds";
import { createAsyncThunk, createSlice, PayloadAction } from "@reduxjs/toolkit";
import { RessuStore, Feeds, RSSChannel } from "./typing";
import isURL from "validator/es/lib/isURL";
import { parseRSSFromURL } from "@/utils/wasm";
import { toast } from "react-hot-toast";
import { remove as lodashRemove } from "lodash-es";
import { concat, uniqWith } from "lodash-es";
import { doubleArrayLoopCall } from "@/utils";

const initialState: Feeds = {
	channels: [],
	filterKeyword: ""
};

/**
 * async add new rss channel.
 */
export const addRSSChannelAsync = createAsyncThunk(
	"channels/add",
	async (url: string) => {
		if (isURL(url)) {
			return await parseRSSFromURL(url);
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
	async (channels: RSSChannel[], { dispatch }): Promise<(RSSChannel | undefined)[]> => {
		dispatch(syncMark);
		const promiseResults: Promise<RSSChannel | undefined>[] = channels.map(
			async (ch): Promise<RSSChannel | undefined> => {
				const result = await parseRSSFromURL(ch.atomLink);
				return result;
			}
		);

		return await Promise.all(promiseResults);
	}
);

export const initinalizeChannelsFromLocal = createAsyncThunk(
	"channels/initFromLocal",
	async (): Promise<RSSChannel[]> => {
		let feedsLocal = await feedsDataLocalGet();
		console.log(feedsLocal);
		if (feedsLocal === null) {
			console.log("init: no local data.");
			await feedsDataLocalSync(initialState);
			feedsLocal = initialState;
		}
		console.log("init from = ", feedsLocal);
		return feedsLocal.channels;
	}
);

const feedsSlice = createSlice({
	name: "channels",
	initialState,
	reducers: {
		initialize: (state: Feeds, { payload }: PayloadAction<Feeds>) => {
			state.channels = payload.channels;
		},

		remove: (state: Feeds, url: PayloadAction<string>) => {
			const { channels } = state;
			lodashRemove(channels, (ch: RSSChannel) => ch.url === url.payload);
		},

		keyword: (state: Feeds, { payload }: PayloadAction<string>) => {
			state.filterKeyword = payload;
		},

		syncMark: (state: Feeds, { payload }: PayloadAction<RSSChannel[]>) => {
			const { channels } = state;
			doubleArrayLoopCall(channels, payload, (a, b) => {
				if (a.atomLink === b.atomLink) {
					console.log("sync mark = " + a.atomLink);
					a.synced = false;
				}
			})
		}
	},

	extraReducers: builder => {
		builder.addCase(
			addRSSChannelAsync.fulfilled,
			(state: Feeds, { payload: channel }) => {
				console.log(channel, state);
				if (channel) {
					if (state.channels.map(t => t.atomLink).includes(channel.atomLink)) {
						toast.error("该订阅源已存在");
						return;
					}

					state.channels.push(channel);
					return state;
				} else {
					toast.error("无效的RSS订阅信息。");
				}
			}
		);

		builder.addCase(
			pullRSSChannelAsync.fulfilled,
			(state: Feeds, { payload: newCh }) => {
				const { channels } = state;
				doubleArrayLoopCall<RSSChannel | undefined>(channels, newCh, (a, b) => {
					if (b === undefined) return;
					if (b!.atomLink === a!.atomLink) {
						a!.posts = concat(b!.posts, a!.posts);
						a!.posts = uniqWith(
							a!.posts,
							(a, b) => a.guid === b.guid && a.title === b.title
						);
						console.log("posts = ", a?.atomLink, a?.posts);
						a!.synced = true;
					}
				});

				state.channels = channels;
				return state;
			}
		);

		builder.addCase(
			initinalizeChannelsFromLocal.fulfilled,
			(state, { payload }: PayloadAction<RSSChannel[]>) => {
				state.channels = payload || [];
			}
		);
	}
});

export const selectFeeds = (state: RessuStore) => state.feeds;

/**
 * It takes the state of the store and returns a filtered list of channels
 * @param {RessuStore} state - RessuStore
 * @returns A function that takes a state and returns a filtered array of channels.
 */
export const selectFeedsByKeyword = (state: RessuStore) => {
	const { filterKeyword, channels } = state.feeds;
	return channels.filter(ch =>
		ch.title
			.toLocaleLowerCase()
			.trim()
			.includes(filterKeyword.toLocaleLowerCase().trim())
	);
};

export const selectChannels = (state: RessuStore) => state.feeds.channels;

export const selectChannelByIndex = (index: number) => (state: RessuStore) =>
	state.feeds.channels[index];

export const selectChannelPostsByIndex =
	(index: number | undefined) => (state: RessuStore) => {
		if (index === undefined || isNaN(index)) {
			return [];
		}
		return state.feeds.channels[index!].posts;
	};

export const selectChannelLength = (state: RessuStore) =>
	state.feeds.channels.length;

export const { initialize, remove, keyword, syncMark } = feedsSlice.actions;

export default feedsSlice.reducer;
