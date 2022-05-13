import { useState } from "react";
import { NavbarContainer } from "./styled";
import FeedAddit from "@/components/FeedAddit/index";
import { Pane, IconButton, PlusIcon, CogIcon } from "evergreen-ui";
import isURL from "validator/es/lib/isURL";
import { useDispatch } from "react-redux";
import { addRSSChannelAsync } from "@store/feeds";
import { infoDelay } from "@/utils/noitce";
import { AppDispath } from "@store/typing";

export default function NavBar() {
	const [additOpen, setAdditOpen] = useState(false);

	const dispatch = useDispatch<AppDispath>();

	const handleOpenAddit = () => {
		setAdditOpen(true);
	};

	const handleCloseAddit = () => {
		setAdditOpen(false);
	}
		const handleSubmitAddit = (url: string) => {
		console.log(url);
		setAdditOpen(false);
		if (isURL(url)) {
			infoDelay("添加：" + url);
			dispatch(addRSSChannelAsync(url));
		} else {
			infoDelay("请输入有效的RSS订阅地址");
		}
	};

	return (
		<NavbarContainer>
			<Pane display="flex" padding={4} justifyContent="space-between">
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
		</NavbarContainer>
	);
}
