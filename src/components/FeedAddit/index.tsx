import { ChangeEvent, useState } from "react";
import PopupPanel from "@/components/PopupPanel"

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
				<PopupPanel open={open} confirm>
				</PopupPanel>
		</>
	);
};

export default FeedAddit;
