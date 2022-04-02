import { ReactNode } from "react";
import { Frame, Container } from "./styled";

type PropsType = {
	open: boolean;
	children?: ReactNode;
};

export default (props: PropsType) => {
	const { open, children } = props;

	if (open) {
		return (
				<Container borderColor={"#DDD"}>{children}</Container>
		);
	}

	return <></>;
};
