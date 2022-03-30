import * as React from "react";
import RessuInput from "@/components/RessuInput";
import classes from "./index.module.scss";
import SearchIcon from "@mui/icons-material/Search"

export default class FeedSidebar extends React.Component {
	constructor(props: any) {
		super(props);
	}

	render(): React.ReactNode {
		return (
			<div className={classes["sidebar-container"]}>
				<RessuInput children={<SearchIcon></SearchIcon>}></RessuInput>
			</div>
		);
	}
}
