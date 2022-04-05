import { ThunkAction } from "redux-thunk";

export type StoreAction = {
    type: Symbol,

    payload?: any
};

export type AsyncAction = ThunkAction<Promise<void>, RessuStore, unknown, StoreAction>;

export type RessuStore = {
    channels: RSSChannel[]

    settings: {},

	hasInitWASM: boolean,
};

export type RSSChannel = {
    title: string,

    url: string,

    description: string,

    posts: Post[]
};


export type Post = {
    title: string,
    url: string,

    // timestrap
    date: string,

    content: string,
};
