import * as React from "react";
import { SearchInput, Icon } from "evergreen-ui";
import { useSelector } from "react-redux";

import { keyword, selectFeedsByKeyword } from "@store/feeds";
import { selectChannel, selectChannelIndex } from "@store/ui-state";
import {
	SidebarContainer,
	SearchInputContainer,
	ChannelItem,
	ChannelItemsContainer,
	ChannelItemText,
	ChannelItemLoading,
	rotateNinjaIcon,
	ChannelItemSelectedBar
} from "./styled";
import { useDispatch } from "react-redux";
import { AppDispath } from "@store/typing";

export default function FeedSidebar() {
	const dispatch = useDispatch<AppDispath>();
	const channels = useSelector(selectFeedsByKeyword);
	const currentChannelIndex = useSelector(selectChannelIndex);

	const channelsJSX = channels.map((t, i) => (
		<ChannelItem key={t.url} onClick={() => handleChannelItemSelectChange(i)}>
			<ChannelItemSelectedBar
				open={currentChannelIndex === i}
			></ChannelItemSelectedBar>

			<ChannelItemText selected={currentChannelIndex === i}>
				{t.title}
			</ChannelItemText>

			{t.synced ? undefined : (
				<ChannelItemLoading>
					<Icon icon={rotateNinjaIcon} size={13} height={"100%"} />
				</ChannelItemLoading>
			)}
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
