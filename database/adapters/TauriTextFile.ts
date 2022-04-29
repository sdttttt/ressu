import { Adapter } from "lowdb/lib/Low";
import { writeFile, readTextFile } from "@tauri-apps/api/fs";

export class TauriTextFile implements Adapter<string> {
	#filePath: string;

	constructor(filePath: string) {
		this.#filePath = filePath;
	}

	read(): Promise<string> {
		return readTextFile(this.#filePath);
	}

	write(data: string): Promise<void> {
		return writeFile({ path: this.#filePath, contents: data });
	}
}
