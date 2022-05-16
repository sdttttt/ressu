import * as React from "react";
import { SearchInput } from "evergreen-ui";
import { keyword, selectFeedsByKeyword } from "@store/feeds";
import { selectChannel } from "@store/ui-state";
import { useSelector } from "react-redux";
import {
	SidebarContainer,
	SearchInputContainer,
	ChannelItem,
	ChannelItemsContainer
} from "./styled";
import { useDispatch } from "react-redux";
import { AppDispath } from "@store/typing";

export default function FeedSidebar() {
	const dispatch = useDispatch<AppDispath>();
	const channels = useSelector(selectFeedsByKeyword);

	const channelsJSX = channels.map((t, i) => (
		<ChannelItem key={t.url} onClick={() => handleChannelItemSelectChange(i)}>
			{" "}
			{t.title}{" "}
		</ChannelItem>
	));

	const handleKeywordChange = (e: string) => {
		dispatch(keyword(e));
	};

	const handleChannelItemSelectChange = (index: number) => {
		console.log("Click on Index: " + index);
		dispatch(selectChannel(index));
	};

	return (
		<SidebarContainer>
			<SearchInputContainer>
				<SearchInput
					width={"95%"}
					onChange={(e: { target: { value: string } }) =>
						handleKeywordChange(e.target.value)
					}
				></SearchInput>
			</SearchInputContainer>

			<ChannelItemsContainer>
				{channels.length === 0 ? <p> 暂无订阅源 </p> : channelsJSX}
			</ChannelItemsContainer>
		</SidebarContainer>
	);
}
