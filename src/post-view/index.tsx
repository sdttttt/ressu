import * as React from "react";

import { selectPostData } from "@store/ui-state";
import { useSelector } from "react-redux";
import { PostViewContainer, PostViewTitle, PostViewTime, PostViewBody } from "./styled";
import { unescape } from "lodash-es";


export default function PostView() {
	const post = useSelector(selectPostData);

	if (!post) {
		return <PostViewContainer> 没有文章 </PostViewContainer>
	}

	return (
		<PostViewContainer>
			<PostViewTitle>{post.title}</PostViewTitle>
			<PostViewTime>{post.pubDate}</PostViewTime>
			<PostViewBody dangerouslySetInnerHTML={{__html: unescape(post.description)}}></PostViewBody>
		</PostViewContainer>
	);
}
