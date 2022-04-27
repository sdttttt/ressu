import * as React from "react";
import RessuInput from "@/components/RessuInput";
import classes from "./index.module.scss";
import { SearchIcon } from "evergreen-ui"
import { selectChannels } from "@store/feeds"
import { useSelector } from "react-redux";

export default function FeedSidebar() {

	const channels = useSelector(selectChannels)


	return (
		<div className={classes["sidebar-container"]}>
			<RessuInput prefix={<SearchIcon />}></RessuInput>

			{ channels.length === 0 ? <h3> 暂无订阅源 </h3> : <></> }
		</div>
	);

}
