
export type StoreAction = {
    type: Symbol,

    payload?: any
};

export type RessuStore = {
    channels: RSSChannel[]

    settings: {},
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