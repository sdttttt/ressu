import type { Settings } from "@store/typing";
import { JSONFile } from "@database/adapters/TauriJSONFile";
import { Low } from "lowdb";
import { DATABASES_PATH, SETTINGS_DB_FILENAME } from "./names";
import { appDir, join } from "@tauri-apps/api/path";

let settingsdb: Low<Settings>;

/**
 * Settings persistence
 * @param data
 * @returns
 */
export async function settingsDataLocalSync(data: Settings) {
	await initializeSettingsDB();
	settingsdb.data = data;
	return await settingsdb.write();
}

/**
 * Get settings from local.
 * @returns
 */
export async function settingsDataLocalGet(): Promise<Settings | null> {
	await initializeSettingsDB();
	await settingsdb.read();
	return settingsdb.data;
}

async function initializeSettingsDB() {
	if (!settingsdb) {
		const rootDir = await appDir();
		settingsdb = new Low(
			new JSONFile<Settings>(
				await join(rootDir, DATABASES_PATH, SETTINGS_DB_FILENAME)
			)
		);
	}
}
