import { ADD_RSS, MARK_READ_ALL_RSS, MARK_READ_RSS, MARK_READ_RSS_ITEM, REMOVE_RSS } from './constants';
import type { StoreAction } from './typing';

export const RSSActions = {
	addRSSAddr: (url: string): StoreAction => ({
		type: ADD_RSS,
		payload: { url },
	}),

	markReadAllRSS: (): StoreAction => ({ type: MARK_READ_ALL_RSS, }),

	markReadRSS: (url: string) => ({
		type: MARK_READ_RSS,
		payload: { url }
	}),

	markReadRSSItem: (url: string, date: string, title: string): StoreAction => ({
		type: MARK_READ_RSS_ITEM,
		payload: { url, date, title }
	}),

	removeRSS: (url: string): StoreAction => ({
		type: REMOVE_RSS,
		payload: { url }
	})
};
