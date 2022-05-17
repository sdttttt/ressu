import { useState } from "react";
import PopupPanel from "@/components/PopupPanel";
import RessuInput from "@/components/RessuInput";
import { FeedIcon } from "evergreen-ui";
import { useAppDispatch } from "@store/index";
import { addRSSChannelAsync } from "@store/feeds";
import { emit } from "@tauri-apps/api/event";
import { RessuEvent } from "@/listens";

type PropsType = {
	open: boolean;
	onClose: () => void;
};

const FeedAddit = (props: PropsType) => {
	const { open, onClose } = props;

	const [url, setURL] = useState("");

	const dispatch = useAppDispatch();

	// 关闭后置空
	const handleOnClose = () => {
		onClose();
		setURL("");
	};

	const handleOnSubmit =  async () => {
		onClose();
		await dispatch(addRSSChannelAsync(url));
		emit(RessuEvent.SyncFeedsToLocal);
	};

	// URL变化
	const handleChangeUrl = (url: string) => {
		setURL(url);
	};

	return (
		<>
			<PopupPanel
				open={open}
				title="订阅源"
				description="请输入需要添加的RSS订阅地址."
				confirm
				onConfirm={handleOnSubmit}
				onCancel={handleOnClose}
			>
				<RessuInput
					prefix={<FeedIcon />}
					onChange={handleChangeUrl}
				></RessuInput>
			</PopupPanel>
		</>
	);
};

export default FeedAddit;
