import { listen } from "@tauri-apps/api/event";


export enum RessuEvent {
	SyncLocal = "ressu://sync-local",
}

const uninstallFnMap = {
	[RessuEvent.SyncLocal]: () => {},
};

export const install = async () => {
	uninstallFnMap[RessuEvent.SyncLocal] =
	await listen(RessuEvent.SyncLocal, async () => {
	});
}

export const uninstall = (eventName: RessuEvent) => {
	uninstallFnMap[eventName]();
}
