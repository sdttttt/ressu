import * as React from "react";
import { InputContainer, InputPrefix, InInput } from "./styled";

type PropsType = {
	onChange?: (e: React.ChangeEvent) => void;
	onInput?: (e: React.FormEvent) => void;
	prefix?: React.ReactNode;
};

const RessuInput = (props: PropsType) => {
	const { prefix, onChange, onInput } = props;

	return (
		<InputContainer>
			{prefix ? (
				<InputPrefix>{prefix}</InputPrefix>
			) : (
				{}
			)}
			<InInput
				onChange={onChange}
				onInput={onInput}
			></InInput>
		</InputContainer>
	);
};

export default RessuInput;
