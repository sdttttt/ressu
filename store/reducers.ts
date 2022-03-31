import { RessuStore, StoreAction, RSSItem } from "./typing";
import { REMOVE_RSS } from "./constants"

const initialState: RessuStore = {
    rss: [],
    settings: {},
};

const rssReducer = (rss: RSSItem[], action: StoreAction): RSSItem[] => {

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


export default ({ rss, settings }: RessuStore = initialState, action: StoreAction) => ({
    rss: rssReducer(rss, action),
    settings: settingsReducer(settings, action)
})
