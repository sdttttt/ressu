import { Low } from "lowdb"
import { JSONFile } from "./adapters/TauriJSONFile"
import { FEED_DB_FILENAME } from "./names";
import { clone } from "lodash-es"

import { Feeds } from "../store/typing"

const feedDB = new Low(new JSONFile<Feeds>(FEED_DB_FILENAME));

/**
 * Feed persistence
 */
export async function feedsDataLocalSync(data: Feeds) {
	const realLocalData = clone(data);
	realLocalData.channels = data.channels.map(
		channel => {
			const result = clone(channel);
			result.posts = [];
			return result;
		}
	)

	feedDB.data = realLocalData;
	return await feedDB.write()
}

/**
 * get Feeds data from local.
 * @returns 
 */
export async function feedsDataLocalGet(): Promise<Feeds> {
	await feedDB.read();
	return feedDB.data;
}

// @ts-ignore
if (import.meta.vitest) {
	const {
		it,
		expect
		// @ts-ignore
	} = import.meta.vitest;

	it('feed copy but not copy posts.', () => {
		const feeds: Feeds = {
			channels: [
				{
					title: "1",
					url: "1.1.1",
					description: "111",
					postSize: 1,
					posts: [{ title: "hellow", url: "hellow", content: "", date: "1" }]
				}
			]
		}

		const realLocalData = clone(feeds);
		realLocalData.channels = feeds.channels.map(
			channel => {
				const result = clone(channel);
				result.posts = [];
				return result;
			}
		)

		expect(realLocalData.channels.length).toBe(1);
		expect(realLocalData.channels[0].title).toBe("1");
		expect(realLocalData.channels[0].url).toBe("1.1.1");
		expect(realLocalData.channels[0].description).toBe("111");
		expect(realLocalData.channels[0].postSize).toBe(1);
		expect(realLocalData.channels[0].posts.length).toBe(0);
	})
}