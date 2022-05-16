import { RSSChannel } from "@store/typing";
import init, { getFeedMeta } from "wasm";
import { toast } from "react-hot-toast";
import { fetchRSSText } from "./http";

const wasmState = {
	ready: false
};

/**
 * It parses RSS feed from text.
 * @param {string} rssText - The text of the RSS feed.
 * @returns It is a promise that resolves to an RSSChannel or undefined.
 */
export const parseRSSFromText = async (
	rssText: string
): Promise<RSSChannel | undefined> => {
	const metaInfo = await runWASM(() => getFeedMeta(rssText));
	if (metaInfo.isSpecification()) {
		return metaInfo.json() as RSSChannel;
	}
	toast.error("Ressu目前只支持RSS2.0.");
};

/**
 * It takes a URL, fetches the RSS text from that URL, and then parses the RSS text into an RSSChannel
 * object
 * @param {string} rssURL - The URL of the RSS feed you want to parse.
 * @returns A Promise that resolves to an RSSChannel or undefined.
 */
export const parseRSSFromURL = async (
	rssURL: string
): Promise<RSSChannel | undefined> => {
	const rssText = await fetchRSSText(rssURL);
	const result = await parseRSSFromText(rssText);

	if (result && !result.atomLink) {
		result.atomLink = rssURL;
	}

	return result;
};

/**
 * run func in wasm runtime.
 */
export const runWASM = async <R = any>(func: () => R): Promise<R> => {
	await initWASM();
	return func();
};

/**
 * generate with args wasm runtime func.
 */
export const runWASMWithArgs = <R = any>(func: (...args: unknown[]) => R) => {
	return async (...arge: unknown[]): Promise<R> => {
		await initWASM();
		return func(...arge);
	};
};

/**
 * init wasm.
 */
export const initWASM = async () => {
	const { ready } = wasmState;
	if (!ready) {
		await init();
		wasmState.ready = true;
	}
};
