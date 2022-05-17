import store from ".";

export type StoreAction = {
	type: Symbol;

	payload?: any;
};

export type RessuStore = ReturnType<typeof store.getState>;

export type AppDispath = typeof store.dispatch;

export type Feeds = {
	channels: RSSChannel[];
	filterKeyword: string;
};

export type RSSChannel = {
	title: string;

	url: string;

	atomLink: string;

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

	synced: boolean,

	copyright?: string;
};

export type Post = {
	title: string;

	link: string;

	guid: string;

	// timestrap
	pubDate: string;

	description: string;

	author?: string;

	category: string[];

	read?: boolean;
};

export type UIState = {
	messages: string[];
	statusBar: {
		type: StatusBarType,
		text: StatusBarText,
	},
	currentChannelIndex?: number;
	currentPostIndex?: number;
};

export enum StatusBarType {
	Buys,
	Free,
};

export enum StatusBarText {
	Free = "当前没有正在执行的任务",

	PullChannel = "正在拉取频道",
	
	SyncLocal = "正在同步本地数据",
};

export type Settings = {};
