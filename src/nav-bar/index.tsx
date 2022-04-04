import { useState } from "react";
import classes from "./index.module.scss";
import FeedAddit from "@/components/FeedAddit/index";
import { Pane, IconButton, PlusIcon, CogIcon, toaster } from "evergreen-ui";
import { successDelay } from "@/utils/noitce";


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
		successDelay("添加：" + url, 1000);
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
