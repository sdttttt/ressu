
export type StoreAction = {
    type: Symbol,

    payload?: any
};

export type RessuStore = {
    rss: RSSItem[]

    settings: {},
};

export type RSSItem = {
    title: string,

    url: string,

    posts: Post[]
};


export type Post = {
    title: string,
    url: string,

    // timestrap
    date: string,

    content: string,
};