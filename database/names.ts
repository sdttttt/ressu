import sha1 from "crypto-js/sha1";
import encUft8 from "crypto-js/enc-utf8";

export const DATABASES_PATH = "databases";

export const FEED_DB_FILENAME = "feeds.json";
export const SETTINGS_DB_FILENAME = "settings.json";

export const POSTS_DB_DIR = "posts";


/**
 * It takes a URL and a date, and returns a filename that is unique to that URL and date
 * @param {string} url - The URL of the post
 * @param {string | Date} date - The date of the post. string: YYYY-mm-dd
 * @returns A string
 */
export const postFileName = (url: string, date: string | Date) => {
	const key = sha1(url).toString(encUft8).slice(0, 16);

	if (date instanceof Date || date.split("-").length !== 3) {
		const dateR = new Date(date);
		date = dateR.getFullYear() + "-" + (dateR.getMonth() + 1) + "-" + dateR.getDate();
	}

	return `${key}-${date}.json`;
};
