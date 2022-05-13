import { Event, EventCallback, listen } from "@tauri-apps/api/event";
import store from "@store/index"
import { feedsDataLocalSync } from "@database/feeds";
import { Feeds, RSSChannel, StatusBarText } from "@store/typing";
import { pullRSSChannelAsync } from "@store/feeds";
import { statusBarBuys, statusBarFree } from "@store/ui-state";

export enum RessuEvent {
	SyncFeedsToLocal = "ressu://sync-feeds-to-local",

	PullAllRSSChannel = "ressu://pull-all-rss-channel"
}

const uninstallFnMap = {
	[RessuEvent.SyncFeedsToLocal]: () => {},
	[RessuEvent.PullAllRSSChannel]: () => {}
};

const handleSyncFeedsToLocal: EventCallback<Feeds | undefined> = async ({ payload }: Event<Feeds | undefined>) => {
	store.dispatch(statusBarBuys(StatusBarText.SyncLocal));
	
	const feeds = payload ? payload : store.getState().feeds;
	await feedsDataLocalSync(feeds);
	
	store.dispatch(statusBarFree(StatusBarText.SyncLocal));
};

const handlePullAllRSSChannel: EventCallback<RSSChannel[] | undefined> = async ({ payload }: Event<RSSChannel[] | undefined>) => {
	store.dispatch(statusBarBuys(StatusBarText.PullChannel));

	const channels = payload ? payload : store.getState().feeds.channels;
	await store.dispatch(pullRSSChannelAsync(channels));

	store.dispatch(statusBarFree(StatusBarText.PullChannel));
}

/**
 * `install` is a function that returns a promise that resolves to a function that, when called, will
 * remove the event listener that was added to the `RessuEvent.SyncLocal` event.
 */
export const install = async () => {
	uninstallFnMap[RessuEvent.SyncFeedsToLocal] = await listen(
		RessuEvent.SyncFeedsToLocal,
		handleSyncFeedsToLocal
	);

	uninstallFnMap[RessuEvent.PullAllRSSChannel] = await listen(
		RessuEvent.PullAllRSSChannel,
		handlePullAllRSSChannel
	);
};

/**
 * It takes an event name as a parameter, and then calls the uninstall function that is mapped to that
 * event name
 * @param {RessuEvent} eventName - The name of the event you want to uninstall.
 */
export const uninstall = (eventName: RessuEvent) => {
	uninstallFnMap[eventName]();
};
