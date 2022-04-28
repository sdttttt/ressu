import { Low } from "lowdb"
import { JSONFile } from "./adapters/TauriJSONFile"
import { FEED_DB_FILENAME } from "./names";
import { clone } from "lodash-es"

import type { Feeds } from "../store/typing"

const feedDB = new Low(new JSONFile<Feeds>(FEED_DB_FILENAME)); 

/**
 * Feed persistence
 */
export async function feedsDataLocalSync(data: Feeds) {
    
}