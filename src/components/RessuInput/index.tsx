import * as React from "react";
import classes from "./index.module.scss";

type PropsType = {
	onChange: (e: React.ChangeEvent) => void;
	onInput: (e: React.FormEvent) => void;
	children?: React.ReactNode;
};

const RessuInput = (props: PropsType) => {
	const { children, onChange, onInput } = props;

	return (
		<div className={classes["input-container"]}>
			{children ? (
				<div className={classes["input-prefix"]}>{children}</div>
			) : (
				""
			)}
			<input
				className={classes["in-input"]}
				onChange={onChange}
				onInput={onInput}
			></input>
		</div>
	);
};

export default RessuInput;
