import type { Post, RSSChannel } from "@store/typing";
import { Low } from "lowdb";
import { JSONFile } from "./adapters/TauriJSONFile";
import sha1 from "crypto-js/sha1";
import encUft8 from "crypto-js/enc-utf8";
import { appDir, sep } from "@tauri-apps/api/path";
import { DATABASES_PATH, POSTS_DB_DIR } from "./names";
import { isString, map } from "lodash-es";

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
		const date = new Date(next.date);
		const dateKey =
			date.getFullYear() + "-" + (date.getMonth() + 1) + "-" + date.getDate();

		if (!total[dateKey]) {
			total[dateKey] = [];
		}

		total[dateKey].push(next);
		return total;
	}, {} as DayPostMap);

	const writeFull = map(dayToPostMap, async (posts, dateKey) => {
		const db = await getPostsDB(channel, new Date(dateKey));
		db.data = posts;
		await db.write();
	});

	// wait for all writes to finish.
	return await Promise.all(writeFull);
}

/**
 * "Get the posts data from the local database for the given channel and date."
 *
 * The first thing we do is check if the date is a string. If it is, we convert it to a Date object
 * @param {RSSChannel} channel - RSSChannel
 * @param {Date | string} date - Date | string
 * @returns The data from the database.
 */
export async function postsDataLocalGet(
	channel: RSSChannel,
	date: Date | string
) {
	if (isString(date)) {
		date = new Date(date);
	}

	const db = await getPostsDB(channel, date);
	await db.read();
	return db.data;
}

/**
 * Since JS does not perform well in reading and writing large files, we use this method to block all articles from * the subscribed source into small files by date.
 *
 * @param param0
 * @param date
 * @returns
 */
async function getPostsDB({ url }: RSSChannel, date: Date) {
	const key = sha1(url).toString(encUft8).slice(0, 16);
	const dateKey =
		date.getFullYear() + "-" + (date.getMonth() + 1) + "-" + date.getDate();

	const rootDir = await appDir();
	const fullFilePath =
		rootDir +
		sep +
		DATABASES_PATH +
		sep +
		POSTS_DB_DIR +
		sep +
		key +
		"-" +
		dateKey +
		".json";

	return new Low(new JSONFile<Post[]>(fullFilePath));
}
