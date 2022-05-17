import * as React from "react";
import {
	TextInput,
	FeedIcon,
	Icon,
	Spinner,
	IconButton,
	PlusIcon,
	SearchIcon,
	Popover,
	Pane,
	RefreshIcon
} from "evergreen-ui";
import { useSelector } from "react-redux";

import { keyword, selectFeedsByKeyword } from "@store/feeds";
import { selectChannel, selectChannelIndex } from "@store/ui-state";
import {
	SidebarContainer,
	TopContainer,
	ChannelItem,
	ChannelItemsContainer,
	ChannelItemText,
	ChannelItemLoading,
	ChannelItemSelectedBar,
	ChannelItemIcon,
	SidebarLabel
} from "./styled";
import { useDispatch } from "react-redux";
import { AppDispath } from "@store/typing";
import { urlRoot } from "@/utils";
import FeedAddit from "@/components/FeedAddit";

export default function FeedSidebar() {
	const [additOpen, setAdditOpen] = React.useState(false);
	const handleOpenAddit = () => {
		setAdditOpen(true);
	};

	const handleCloseAddit = () => {
		setAdditOpen(false);
	};

	const dispatch = useDispatch<AppDispath>();
	const channels = useSelector(selectFeedsByKeyword);
	const currentChannelIndex = useSelector(selectChannelIndex);

	const channelsJSX = channels.map((t, i) => (
		<ChannelItem key={t.url} onClick={() => handleChannelItemSelectChange(i)}>
			<ChannelItemSelectedBar
				open={currentChannelIndex === i}
			></ChannelItemSelectedBar>

			<ChannelItemIcon>
				{t?.image?.url ? (
					<img src={t.image.url} width={"100%"} height={"100%"}></img>
				) : (
					<img src={urlRoot(t.url) + "/favicon.ico"}></img>
				)}
			</ChannelItemIcon>

			<ChannelItemText selected={currentChannelIndex === i}>
				{t.title}
			</ChannelItemText>

			{t.synced ? undefined : (
				<ChannelItemLoading>
					<Spinner size={13} />
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
			<TopContainer>
				<Pane display="flex" flexDirection="row" justifyContent="start">
					<Popover
						bringFocusInside
						content={
							<Pane
								padding={10}
								display="flex"
								alignItems="center"
								justifyContent="center"
								flexDirection="column"
							>
								<TextInput
									width={"100%"}
									placeholder="feed title keywords..."
									onChange={(e: { target: { value: string } }) =>
										handleKeywordChange(e.target.value)
									}
								></TextInput>
							</Pane>
						}
					>
						<IconButton icon={SearchIcon} appearance="minimal" />
					</Popover>
				</Pane>

				<Pane
					display="flex"
					flexShrink={0}
					flexDirection="row"
					justifyContent="end"
				>
					<IconButton icon={PlusIcon} onClick={() => handleOpenAddit()} appearance="minimal" />
					<IconButton icon={RefreshIcon} appearance="minimal" />
				</Pane>
			</TopContainer>

			<SidebarLabel>
				<Icon icon={FeedIcon} marginRight="3px" size={12}></Icon> 订阅源
			</SidebarLabel>

			<ChannelItemsContainer>
				{channels.length === 0 ? <p> 暂无订阅源 </p> : channelsJSX}
			</ChannelItemsContainer>

			<FeedAddit
				open={additOpen}
				onClose={handleCloseAddit}
			></FeedAddit>
		</SidebarContainer>
	);
}
