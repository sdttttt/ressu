import * as React from "react";

import { useSelector } from "react-redux";
import { selectChannelIndex } from "@store/ui-state";
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

export default function FeedItems() {
	const currentChannelIndex = useSelector(selectChannelIndex);

	const posts = useSelector(selectChannelPostsByIndex(currentChannelIndex)).map(
		(t, index) => (
			<PostItem key={t.guid}>
				<PostItemTitle>{t.title}</PostItemTitle>
				<PostItemBody>{filterXSS(unescape(t.description), { stripIgnoreTag: true, stripIgnoreTagBody: ['script'], whiteList: {} })}</PostItemBody>
				<PostItemTime>{t.pubDate}</PostItemTime>
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
