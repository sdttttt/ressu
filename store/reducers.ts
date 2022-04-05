import { RessuStore, StoreAction, RSSChannel } from "./typing";
import { INITIALIZE_WASM, REMOVE_RSS } from "./constants"

const initialState: RessuStore = {
    channels: [],
    settings: {},
	hasInitWASM: false,
};

const rssReducer = (rss: RSSChannel[], action: StoreAction): RSSChannel[] => {

    switch(action.type) {
        case REMOVE_RSS:
            const { url } = action.payload;
            return rss.filter(t => t.url != url);
    }

    return rss;
}

const settingsReducer = (settings: {}, action: StoreAction): {} => {
    return settings;
}

const wasmReducer = (hasInitWASM: boolean, action: StoreAction) => {
	switch(action.type) {
		case INITIALIZE_WASM:
			return true;
	}

	return hasInitWASM;
};

export default ({ channels, settings, hasInitWASM }: RessuStore = initialState, action: StoreAction) => ({
    channels: rssReducer(channels, action),
    settings: settingsReducer(settings, action),
	hasInitWASM: wasmReducer(hasInitWASM, action),
})
