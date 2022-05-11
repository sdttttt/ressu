import store from ".";

export type StoreAction = {
	type: Symbol;

	payload?: any;
};

export type RessuStore = ReturnType<typeof store.getState>;

export type Feeds = {
	channels: RSSChannel[];
};

export type RSSChannel = {
	title: string;

	url: string;

	description: string;

	language?: string;

	webMaster?: string;

	generator?: string;

	lastBuildDate?: string;

	ttl?: number;

	image?: {
		url?: string;
	};

	posts: Post[];

	copyright?: string;
};

export type Post = {
	title: string;

	link: string;

	guid: string;

	// timestrap
	pubDate: string;

	descriptioin: string;

	author?: string;

	category: string[];

	read?: boolean;
};

export type Message = {
	contents: string[];
};

export type Settings = {};
