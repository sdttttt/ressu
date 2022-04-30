import * as React from "react";
import RessuInput from "@/components/RessuInput";
import { SearchIcon } from "evergreen-ui";
import { selectChannels } from "@store/feeds";
import { useSelector } from "react-redux";
import { SidebarContainer } from "./styled"

export default function FeedSidebar() {
	const channels = useSelector(selectChannels);

	const channelsJSX = channels.map(t => <div> {t.title} </div>);

	return (
		<SidebarContainer>
			<RessuInput prefix={<SearchIcon />}></RessuInput>

			{channels.length === 0 ? <h3> 暂无订阅源 </h3> : channelsJSX}
		</SidebarContainer>
	);
}
