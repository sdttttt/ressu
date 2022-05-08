import { createDir } from "@tauri-apps/api/fs";
import { appDir, join } from "@tauri-apps/api/path";
import { DATABASES_PATH } from "@database/names";
import { initinalizeFeedsFromLocal } from "@store/feeds";
import store from "@store/index";
import { install as installListens } from "./listens";
import { fetchRSSText } from "./utils/http";

(async () => {
	installListens();
	const rootPathAsync = await appDir();
	await createDir(await join(rootPathAsync, DATABASES_PATH), {
		recursive: true
	});

	store.dispatch(initinalizeFeedsFromLocal());
})();
