import { ReactNode, MouseEventHandler } from "react";
import { Container, ConfirmDiv } from "./styled";
import Button from "@mui/material/Button";

type PropsType = {
	/**
	 * popup display.
	 */
	open: boolean;

	/**
	 * confirm and cancel button display.
	 */
	confirm?: boolean;
	
	/**
	 * confirm button click callback.
	 */
	onConfirm?: MouseEventHandler<HTMLButtonElement>,
	
	/**
	 * cancel button click callback;   
	 */
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
