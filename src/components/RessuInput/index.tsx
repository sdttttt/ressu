import * as React from "react";
import { InputContainer, InputPrefix, InInput } from "./styled";

type PropsType = {
	onChange?: (_: string) => void;
	prefix?: React.ReactNode;
};

const RessuInput = (props: PropsType) => {
	const { prefix, onChange } = props;

	// URL变化
	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const value = e.target.value;
		onChange ? onChange(value) : null;
	};

	return (
		<InputContainer>
			{prefix ? <InputPrefix>{prefix}</InputPrefix> : {}}
			<InInput onChange={handleChange}></InInput>
		</InputContainer>
	);
};

export default RessuInput;
