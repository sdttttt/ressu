import { createDir } from "@tauri-apps/api/fs";
import store from "@store/index";
import { appDir, join } from "@tauri-apps/api/path";
import { DATABASES_PATH } from "@database/names";
import { initinalizeFeedsFromLocal } from "@store/feeds";

(async () => {
	const rootPathAsync = await appDir();
	await createDir(await join(rootPathAsync, DATABASES_PATH), { recursive: true });
	store.dispatch(initinalizeFeedsFromLocal());
})()


