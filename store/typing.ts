import store from ".";

export type StoreAction = {
    type: Symbol,

    payload?: any
};

export type RessuStore = ReturnType<typeof store.getState>;

export type Feeds = {
    channels: RSSChannel[],
};

export type RSSChannel = {
    title: string,

    url: string,

    description: string,

    postSize: number,

    posts?: Post[]
};


export type Post = {
    title: string,
    url: string,

    // timestrap
    date: string,

    content: string,
};

export type Message = {
    contents: string[]
};

export type Settings = {};