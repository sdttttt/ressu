import { ChangeEvent, useState } from "react";
import PopupPanel from "@/components/PopupPanel";
import RessuInput from "@/components/RessuInput";
import { FeedIcon } from "evergreen-ui";

type PropsType = {
	open: boolean;
	onClose: () => void;
	onSubmit: (url: string) => void;
};

const FeedAddit = (props: PropsType) => {
	const { open, onClose, onSubmit } = props;

	const [url, setURL] = useState("");

	// 关闭后置空
	const handleOnClose = () => {
		onClose();
		setURL("");
	};

	const handleOnSubmit = () => {
		onSubmit(url);
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
