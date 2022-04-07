import { useState } from "react";
import classes from "./index.module.scss";
import FeedAddit from "@/components/FeedAddit/index";
import { Pane, IconButton, PlusIcon, CogIcon } from "evergreen-ui";
import { dangerDelay, successDelay } from "@/utils/noitce";
import isURL from "validator/es/lib/isURL"


export default function NavBar() {

	const [additOpen, setAdditOpen] = useState(false);

	const handleOpenAddit = () => {
		setAdditOpen(true);
	};

	const handleCloseAddit = () => {
		setAdditOpen(false);
	};

	const handleSubmitAddit = (url: string) => {
		console.log(url);
		setAdditOpen(false);
		if (isURL(url)) {
			successDelay("添加：" + url, 300);
		} else {
			dangerDelay("输入的不是一个URL.", 300);
		}
	};

	return (
		<div className={classes["navbar-container"]}>
			<Pane
				display="flex"
				padding={4}
				justifyContent="space-between"
			>
				<Pane display="flex" justifyContent="left" alignItems="center">
					<Pane>
						<IconButton icon={PlusIcon} onClick={handleOpenAddit}></IconButton>
					</Pane>
				</Pane>

				<Pane display="flex" justifyContent="right" alignItems="center">
					<Pane>
						<IconButton icon={CogIcon}></IconButton>
					</Pane>
				</Pane>
			</Pane>

			<FeedAddit
				open={additOpen}
				onClose={handleCloseAddit}
				onSubmit={handleSubmitAddit}
			></FeedAddit>
		</div>
	);
}
