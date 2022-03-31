import { RessuStore, StoreAction, RSSItem } from "./typing";

const initialState: RessuStore = {
    rss: [],
    settings: {},
};

const rssReducer = (rss: RSSItem[], action: StoreAction): RSSItem[] => {
    return rss;
}

const settingsReducer = (settings: {}, action: StoreAction): {} => {
    return settings;
}


export default ({ rss, settings }: RessuStore = initialState, action: StoreAction) => ({
    rss: rssReducer(rss, action),
    settings: settingsReducer(settings, action)
})
