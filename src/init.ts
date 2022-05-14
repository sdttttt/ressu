import { createDir } from "@tauri-apps/api/fs";
import { appDir, join } from "@tauri-apps/api/path";
import { DATABASES_PATH } from "@database/names";
import { initinalizeChannelsFromLocal } from "@store/feeds";
import store from "@store/index";
import * as Listens from "./listens";

(async () => {
	Listens.install();
	const rootPath = await appDir();
	await createDir(await join(rootPath, DATABASES_PATH), {
		recursive: true
	});

	await store.dispatch(initinalizeChannelsFromLocal());
})();
