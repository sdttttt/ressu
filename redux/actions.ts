import { ADD_RSS, MARK_READ_ALL_RSS, MARK_READ_RSS, MARK_READ_RSS_ITEM } from './constants';

export function addRSSAddr(url: string) {
	return {
		type: ADD_RSS,
		payload: { url },
	};
}

export function markReadAllRSS() {
	return {
		type: MARK_READ_ALL_RSS,
	}
}

export function markReadRSS(url: string) {
	return {
		type: MARK_READ_RSS,
		payload: { url }
	}
}

export function markReadRSSItem(url: string, date: string, title: string) {
	return {
		type: MARK_READ_RSS_ITEM,
		payload: { url, date, title }
	}
}