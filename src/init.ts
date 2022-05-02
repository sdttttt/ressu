import { createDir } from "@tauri-apps/api/fs";
import { appDir, join } from "@tauri-apps/api/path";
import { DATABASES_PATH } from "@database/names";

const rootPathAsync = await appDir();

await createDir(await join(rootPathAsync, DATABASES_PATH), { recursive: true });
