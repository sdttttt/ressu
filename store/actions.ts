import { ADD_RSS_CHANNEL, MARK_READ_ALL_RSS_CHANNEL, MARK_READ_RSS_CHANNEL, MARK_READ_RSS_POST, REMOVE_RSS } from './constants';
import type { StoreAction } from './typing';

export const RSSActions = {
	addRSSAddr: (url: string): StoreAction => ({
		type: ADD_RSS_CHANNEL,
		payload: { url },
	}),

	markReadAllRSS: (): StoreAction => ({ type: MARK_READ_ALL_RSS_CHANNEL, }),

	markReadRSS: (url: string) => ({
		type: MARK_READ_RSS_CHANNEL,
		payload: { url }
	}),

	markReadRSSItem: (url: string, date: string, title: string): StoreAction => ({
		type: MARK_READ_RSS_POST,
		payload: { url, date, title }
	}),

	removeRSS: (url: string): StoreAction => ({
		type: REMOVE_RSS,
		payload: { url }
	})
};
