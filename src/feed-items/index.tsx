import * as React from "react";
import { useAppDispatch } from "@store/index";

import { useSelector } from "react-redux";
import { selectChannelIndex, selectPost, selectPostIndex } from "@store/ui-state";
import { selectChannelPostsByIndex } from "@store/feeds";
import {
	PostItemContainer,
	PostItem,
	PostItemBody,
	PostItemTime,
	PostItemTitle
} from "./styled";
import { filterXSS } from "xss"
import { unescape } from "lodash-es";
import { formatDateFromStr } from "@/utils";

export default function FeedItems() {
	const currentChannelIndex = useSelector(selectChannelIndex);

	const dispath = useAppDispatch();

	const currentPostIndex = useSelector(selectPostIndex);

	const handlePostSelected = (index: number) => {
		console.log("Click on Post Index: " + index);
		dispath(selectPost(index));
	}

	const posts = useSelector(selectChannelPostsByIndex(currentChannelIndex)).map(
		(t, index) => (
			<PostItem key={t.guid} onClick={() => handlePostSelected(index)} selected={currentPostIndex === index}>
				<PostItemTitle>{t.title}</PostItemTitle>
				<PostItemBody>{filterXSS(unescape(t.description), { stripIgnoreTag: true, stripIgnoreTagBody: ['script'], whiteList: {} })}</PostItemBody>
				<PostItemTime>{ formatDateFromStr(t.pubDate)}</PostItemTime>
			</PostItem>
		)
	);

	return (
		<>
			{posts.length === 0 ? (
				<p>暂时没有文章哦 ~</p>
			) : (
				<PostItemContainer>{posts}</PostItemContainer>
			)}
		</>
	);
}
