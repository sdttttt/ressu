import * as React from "react";
import { useSelector } from "react-redux";
import { selectChannelIndex } from "@store/ui-state";
import { selectChannelPostsByIndex } from "@store/feeds";
import { PostItem, PostItemBody, PostItemTitle } from "./styled";

export default function FeedItems(props: any) {
	const currentChannelIndex = useSelector(selectChannelIndex);
	const posts = useSelector(selectChannelPostsByIndex(currentChannelIndex)).map(
		t => (
			<PostItem>
				<PostItemTitle>{t.title}</PostItemTitle>
				<PostItemBody>{t.description}</PostItemBody>
			</PostItem>
		)
	);

	return <>{posts.length === 0 ? <p>暂时没有文章哦 ~</p> : posts}</>;
}
