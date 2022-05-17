import * as React from "react";
import { SearchInput, FeedIcon, Icon } from "evergreen-ui";
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
	ChannelItemSelectedBar,
	ChannelItemIcon,
	SidebarLabel
} from "./styled";
import { useDispatch } from "react-redux";
import { AppDispath } from "@store/typing";
import { urlRoot } from "@/utils";

export default function FeedSidebar() {
	const dispatch = useDispatch<AppDispath>();
	const channels = useSelector(selectFeedsByKeyword);
	const currentChannelIndex = useSelector(selectChannelIndex);

	const channelsJSX = channels.map((t, i) => (
		<ChannelItem key={t.url} onClick={() => handleChannelItemSelectChange(i)}>
			<ChannelItemSelectedBar
				open={currentChannelIndex === i}
			></ChannelItemSelectedBar>

			<ChannelItemIcon>
				{
					t?.image?.url ? 
						<img src={t.image.url} width={"100%"} height={"100%"}></img>
					: <img src={ urlRoot(t.url) + "/favicon.ico"}></img>
				}
			</ChannelItemIcon>

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

			<SidebarLabel>
				<Icon icon={FeedIcon} marginRight="3px" size={12}></Icon> 订阅源
			</SidebarLabel>

			<ChannelItemsContainer>
				{channels.length === 0 ? <p> 暂无订阅源 </p> : channelsJSX}
			</ChannelItemsContainer>
		</SidebarContainer>
	);
}
