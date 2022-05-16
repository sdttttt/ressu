import { ReactNode, MouseEventHandler } from "react";
import {
	Frame,
	Container,
	ConfirmDiv,
	HeaderDiv,
	TitleH4,
	DescriptionSpan
} from "./styled";
import { Button, Overlay } from "evergreen-ui";

type PropsType = {
	/**
	 * popup display.
	 */
	open: boolean;

	/**
	 * title text.
	 */
	title?: string;

	/**
	 * description text.
	 */
	description?: string;

	/**
	 * confirm and cancel button display.
	 */
	confirm?: boolean;

	/**
	 * confirm button click callback.
	 */
	onConfirm?: MouseEventHandler<HTMLButtonElement>;

	/**
	 * cancel button click callback;
	 */
	onCancel: () => void;

	children?: ReactNode;
};

export default (props: PropsType) => {
	const { open, confirm, title, description, onConfirm, onCancel, children } =
		props;

	const HeaderBar = () => (
		<HeaderDiv>
			{title ? <TitleH4>{title}</TitleH4> : ""}
			{description ? <DescriptionSpan>{description}</DescriptionSpan> : ""}
		</HeaderDiv>
	);

	const ConfirmBar = () => (
		<ConfirmDiv>
			<div>
				<Button appearance="minimal" onClick={onCancel}>
					取消
				</Button>
			</div>

			<div>
				<Button appearance="minimal" intent="success" onClick={onConfirm}>
					确认
				</Button>
			</div>
		</ConfirmDiv>
	);

	return (
		<>
			<Overlay
				isShown={open}
				onBeforeClose={() => {
					onCancel();
					return true;
				}}
				children={""}
			/>

			<Container open={open} borderColor={"#DDD"}>
				<HeaderBar></HeaderBar>

				<Frame>{children}</Frame>

				{confirm ? <ConfirmBar /> : ""}
			</Container>
		</>
	);
};
