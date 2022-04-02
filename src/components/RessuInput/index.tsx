import * as React from "react";
import classes from "./index.module.scss";

type PropsType = {
	onChange?: (e: React.ChangeEvent) => void;
	onInput?: (e: React.FormEvent) => void;
	prefix?: React.ReactNode;
};

const RessuInput = (props: PropsType) => {
	const { prefix, onChange, onInput } = props;

	return (
		<div className={classes["input-container"]}>
			{prefix ? (
				<div className={classes["input-prefix"]}>{prefix}</div>
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
