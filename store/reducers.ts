import { RessuStore, StoreAction, RSSChannel } from "./typing";
import { REMOVE_RSS } from "./constants"

const initialState: RessuStore = {
    channels: [],
    settings: {},
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


export default ({ channels, settings }: RessuStore = initialState, action: StoreAction) => ({
    channels: rssReducer(channels, action),
    settings: settingsReducer(settings, action)
})
