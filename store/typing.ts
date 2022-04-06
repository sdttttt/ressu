import store from ".";

export type StoreAction = {
    type: Symbol,

    payload?: any
};

export type RessuStore = ReturnType<typeof store.getState>;


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

export type Settings = {};