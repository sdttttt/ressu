import { Adapter } from "lowdb/lib/Low";
import { writeFile, readTextFile, readDir } from "@tauri-apps/api/fs";
import { sep } from "@tauri-apps/api/path";
import { chain } from "lodash-es";

export class TauriTextFile implements Adapter<string> {
	#filePath: string;

	init: boolean;

	constructor(filePath: string) {
		this.#filePath = filePath;
		this.init = false;
	}

	async read(): Promise<string> {
		if (!this.init) {
			const dirPath = this.#filePath
				.split(sep)
				.slice(0, this.#filePath.split(sep).length - 1)
				.join(sep);

			const filename = this.#filePath.split(sep).pop();

			const files = await readDir(dirPath);

			const exist = files.map(file => file.name).includes(filename);

			if (!exist) {
				writeFile({ path: this.#filePath, contents: "" });
			}

			this.init = true;
		}

		return readTextFile(this.#filePath);
	}

	write(data: string): Promise<void> {
		return writeFile({ path: this.#filePath, contents: data });
	}
}
