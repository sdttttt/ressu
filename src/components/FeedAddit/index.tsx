import { ChangeEvent, useState } from "react";
import PopupPanel from "@/components/PopupPanel"
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
		setURL("");
	};

	// URL变化
	const handleChangUrl = (e: ChangeEvent<HTMLInputElement>) => {
		const url = e.target.value;
		setURL(url);
	};

	return (
		<>
			<PopupPanel
				open={open}
				title="订阅源"
				description="请输入需要添加的RSS订阅地址."
				confirm
			>

				<RessuInput prefix={<FeedIcon />}></RessuInput>
			</PopupPanel>
		</>
	);
};

export default FeedAddit;
