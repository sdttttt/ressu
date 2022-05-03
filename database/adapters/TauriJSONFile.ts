import { Adapter } from "lowdb/lib/Low";
import { TauriTextFile } from "./TauriTextFile";

export class JSONFile<T> implements Adapter<T> {
	#adapter: TauriTextFile;

	constructor(filename: string) {
		this.#adapter = new TauriTextFile(filename);
	}

	async read(): Promise<T | null> {
		const data = await this.#adapter.read();
		if (data === null) {
			return null;
		} else {
			try {
				return JSON.parse(data) as T;
			} catch(_) {
				return null;
			}
		}
	}

	write(obj: T): Promise<void> {
		return this.#adapter.write(JSON.stringify(obj, null, 2));
	}
}
