import type { Post, RSSChannel } from "@store/typing";
import { Low } from "lowdb";
import { JSONFile } from "./adapters/TauriJSONFile";
import { appDir, join } from "@tauri-apps/api/path";
import { DATABASES_PATH, postFileName, POSTS_DB_DIR } from "./names";
import { map, uniqWith, concat } from "lodash-es";

type DayPostMap = {
	[key: string]: Post[];
};

/**
 * It takes a channel, and then it takes all the posts in that channel, and then it groups them by
 * date, and then it writes them to the database
 * @param {RSSChannel} channel - RSSChannel
 * @returns An array of promises.
 */
export async function postDataLocalSync(channel: RSSChannel) {
	const { posts } = channel;

	if (!posts) return;

	const dayToPostMap = posts.reduce((total, next) => {
		const date = new Date(next.pubDate);
		const dateKey =
			date.getFullYear() + "-" + (date.getMonth() + 1) + "-" + date.getDate();

		if (!total[dateKey]) {
			total[dateKey] = [];
		}

		total[dateKey].push(next);
		return total;
	}, {} as DayPostMap);

	const writeFull = map(dayToPostMap, async (posts, dateKey) => {
		const db = await getPostsDB(channel, dateKey);
		db.read();
		const mergeData = concat(db.data || [], posts);
		db.data = uniqWith(
			mergeData,
			(a, b) => a.guid === b.guid && a.title === b.title
		);

		await db.write();
	});

	// wait for all writes to finish.
	return await Promise.all(writeFull);
}

/**
 * "Get the posts data from the local database for the given channel and date."
 *
 * @param {RSSChannel} channel - RSSChannel
 * @param {Date | string} date - Date | string
 * @returns The data from the database.
 */
export async function postsDataLocalGet(
	channel: RSSChannel,
	date: Date | string
) {
	const db = await getPostsDB(channel, date);
	await db.read();
	return db.data;
}

/**
 * It returns a database of posts for a given RSS channel and date.
 * @param {RSSChannel}  - RSSChannel - is an object that contains the url of the RSS feed
 * @param {string | Date} date - string | Date
 * @returns A new instance of the LowDB class.
 */
async function getPostsDB({ url }: RSSChannel, date: string | Date) {
	const rootDir = await appDir();
	const filename = postFileName(url, date);

	const fullFilePath = await join(
		rootDir,
		DATABASES_PATH,
		POSTS_DB_DIR,
		filename
	);
	return new Low(new JSONFile<Post[]>(fullFilePath));
}
