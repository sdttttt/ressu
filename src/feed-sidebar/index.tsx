import * as React from "react";
import RessuInput from "@/components/RessuInput";
import { SearchIcon } from "evergreen-ui";
import { keyword, selectFeedsByKeyword } from "@store/feeds";
import { useSelector } from "react-redux";
import { SidebarContainer } from "./styled";
import { useDispatch } from "react-redux"
import { AppDispath } from "@store/typing";

export default function FeedSidebar() {
	const dispatch = useDispatch<AppDispath>();
	const channels = useSelector(selectFeedsByKeyword);

	const channelsJSX = channels.map(t => <div> {t.title} </div>);

	const handleKeywordChange = (e: string) => {
		dispatch(keyword(e));
	}

	return (
		<SidebarContainer>
			<RessuInput onChange={handleKeywordChange}  prefix={<SearchIcon />}></RessuInput>

			{channels.length === 0 ? <h3> 暂无订阅源 </h3> : channelsJSX}
		</SidebarContainer>
	);
}
