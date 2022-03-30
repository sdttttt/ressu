import { ADD_RSS_ADDRESS } from './constants';

export function addRSSAddr(url: string) {
	return {
		type: ADD_RSS_ADDRESS,
		payload: url,
	};
}