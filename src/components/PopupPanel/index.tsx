import { ReactNode, EventHandler, MouseEventHandler } from "react";
import { Container, ConfirmDiv } from "./styled";
import Button from "@mui/material/Button";

type PropsType = {
	open: boolean;
	confirm?: boolean;
	onConfirm?: MouseEventHandler<HTMLButtonElement>,
	onCancel?: MouseEventHandler<HTMLButtonElement>,
	children?: ReactNode;
};

export default (props: PropsType) => {
	const { open, confirm,
		onConfirm, onCancel,
		children } = props;

	const ConfirmBar = () => (
		<ConfirmDiv>
			<div>
				<Button size="small" onClick={onCancel}>
					取消
				</Button>
			</div>

			<div>
				<Button color="success" size="small" onClick={onConfirm}>
					确认
				</Button>
			</div>
		</ConfirmDiv>
	);

	return (
		<Container open={open} borderColor={"#DDD"}>
			{children}
			{confirm ? <ConfirmBar /> : ""}
		</Container>
	);


};
