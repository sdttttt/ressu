import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogTitle from "@mui/material/DialogTitle";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import { ChangeEvent, useState } from "react";

type PropsType = {
	open: boolean;
	onClose: () => void;
	onSubmit: (url: string) => void;
};

const FeedAddit = (props: PropsType) => {

	const [url, setURL] = useState("");

	const handleChangUrl = (e: ChangeEvent<HTMLInputElement>) => {
		const url = e.target.value;
		console.log(url);
		setURL(url);
	};

	return (
		<>
			<Dialog open={props.open} onClose={props.onClose} fullWidth>
				<DialogTitle>添加RSS源</DialogTitle>
				<DialogContent>
					<DialogContentText>输入RSS源订阅的地址</DialogContentText>
					<TextField
						autoFocus
						margin="dense"
						id="name"
						label="URL"
						type="text"
						fullWidth
						variant="standard"
						onChange={handleChangUrl}
					/>
				</DialogContent>
				<DialogActions>
					<Button onClick={props.onClose}>取消</Button>
					<Button onClick={() => props.onSubmit(url)}>订阅</Button>
				</DialogActions>
			</Dialog>
		</>
	);
};

export default FeedAddit;
