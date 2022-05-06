import { listen } from "@tauri-apps/api/event";

export enum RessuEvent {
	SyncLocal = "ressu://sync-local",

	PullAllRSS = "ressu://pull-all-rss"
}

const uninstallFnMap = {
	[RessuEvent.SyncLocal]: () => {},
	[RessuEvent.PullAllRSS]: () => {}
};

/**
 * `install` is a function that returns a promise that resolves to a function that, when called, will
 * remove the event listener that was added to the `RessuEvent.SyncLocal` event.
 */
export const install = async () => {
	uninstallFnMap[RessuEvent.SyncLocal] = await listen(
		RessuEvent.SyncLocal,
		async () => {}
	);

	uninstallFnMap[RessuEvent.PullAllRSS] = await listen(
		RessuEvent.PullAllRSS,
		async () => {}
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
