import { createDir } from "@tauri-apps/api/fs";
import { appDir, join } from "@tauri-apps/api/path";
import { DATABASES_PATH } from "@database/names";
import { initinalizeChannelsFromLocal, pullRSSChannelAsync } from "@store/feeds";
import store from "@store/index";
import * as Listens from "./listens";
import { emit } from "@tauri-apps/api/event";

(async () => {
	Listens.install();
	const rootPath = await appDir();
	await createDir(await join(rootPath, DATABASES_PATH), {
		recursive: true
	});

	await store.dispatch(initinalizeChannelsFromLocal());
	emit(Listens.RessuEvent.PullAllRSSChannel);
})();
