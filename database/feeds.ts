import type { Feeds, RSSChannel } from "@store/typing";
import { Low } from "lowdb";
import { JSONFile } from "./adapters/TauriJSONFile";
import { DATABASES_PATH, FEED_DB_FILENAME } from "./names";
import { clone } from "lodash-es";
import { appDir, join } from "@tauri-apps/api/path";

let feedDB: Low<Feeds>;

/**
 * It takes a `Feeds` object, clones it, and then sets the `posts` property of each `Channel` object to
 * an empty array.
 * @param {Feeds} data - Feeds
 */
export async function feedsDataLocalSync(data: Feeds) {
	await initializeFeedDB();
	const realLocalData = clone(data);
	realLocalData.channels = data.channels.map(channel => {
		const result = clone(channel);
		result.posts = [];
		return result;
	});

	feedDB.data = realLocalData;
	console.log("Sync Feeds Data:", realLocalData);
	return await feedDB.write();
}

/**
 * It gets the data from the local database
 * @returns A promise that resolves to the data in the feedDB.
 */
export async function feedsDataLocalGet(): Promise<Feeds | null> {
	await initializeFeedDB();
	await feedDB.read();
	return feedDB.data;
}

/**
 * It adds a channel to the local database
 * @param {RSSChannel} channel - RSSChannel - The channel to add to the database
 * @returns The promise of the write operation.
 */
export async function addChannelDataLocalSync(channel: RSSChannel) {
	await initializeFeedDB();
	await feedDB.read();
	if (!feedDB.data) {
		throw new Error("No data in the feedDB");
	}
	feedDB.data.channels.push(channel);
	return await feedDB.write();
}

/**
 * It creates a new instance of the LowDB database, and it uses a JSON file as the storage medium
 */
async function initializeFeedDB() {
	if (!feedDB) {
		const rootDir = await appDir();
		console.log("root dir = ", rootDir);
		feedDB = new Low(
			new JSONFile<Feeds>(await join(rootDir, DATABASES_PATH, FEED_DB_FILENAME))
		);
	}
}

if (import.meta.vitest) {
	const { it, expect } = import.meta.vitest;

	it("feed copy but not copy posts.", () => {
		const feeds: Feeds = {
			channels: [
				{
					title: "1",
					url: "1.1.1",
					description: "111",
					atomLink: "111.1",
					posts: [{ title: "hellow", link: "hellow", pubDate: "", description: "1", guid: "hellow", category:[] }]
				}
			],
			filterKeyword: ""
		};

		const realLocalData = clone(feeds);
		realLocalData.channels = feeds.channels.map(channel => {
			const result = clone(channel);
			result.posts = [];
			return result;
		});

		expect(realLocalData.channels.length).toBe(1);
		expect(realLocalData.channels[0].title).toBe("1");
		expect(realLocalData.channels[0].url).toBe("1.1.1");
		expect(realLocalData.channels[0].description).toBe("111");
		expect(realLocalData.channels[0]!.posts!.length).toBe(0);
	});
}
