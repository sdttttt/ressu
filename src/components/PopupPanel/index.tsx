import { ReactNode } from "react";
import { Container, ConfirmDiv } from "./styled";
import Button from "@mui/material/Button";

type PropsType = {
	open: boolean;
	confirm?: boolean;
	onConfirm?: (e: Event ) => void
	children?: ReactNode;
};

export default (props: PropsType) => {
	const { open, confirm, children } = props;

	const ConfirmBar = () => (
		<ConfirmDiv>
			<div>
				<Button  size="small">
					取消
				</Button>
			</div>

			<div>
				<Button color="success"  size="small">
					确认
				</Button>
			</div>
		</ConfirmDiv>
	);

	if (open) {
		return (
			<Container borderColor={"#DDD"}>
				{children}
				{confirm ? <ConfirmBar /> : ""}
			</Container>
		);
	}

	return <></>;
};
