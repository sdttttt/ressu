import { useState } from "react";
import classes from "./index.module.scss";
import { Grid, IconButton } from "@mui/material";
import SettingsIcon from "@mui/icons-material/Settings";
import AddIcon from "@mui/icons-material/Add";
import FeedAddit from "@/components/FeedAddit/index";
import { useSnackbar } from "notistack";

export default function NavBar() {
	const { enqueueSnackbar } = useSnackbar();

	const [additOpen, setAdditOpen] = useState(false);

	const handleOpenAddit = () => {
		setAdditOpen(true);
	};

	const handleCloseAddit = () => {
		setAdditOpen(false);
	};

	const handleSubmitAddit = (url: string) => {
		console.log(url);
		setAdditOpen(false);
		enqueueSnackbar("URL: " + url, { variant: "success" });
	};

	return (
		<div className={classes["navbar-container"]}>
			<Grid container columns={24}>
				<Grid md={12} xs={12} columns={24} container item>
					<Grid item md={3} xs={3} textAlign="center">
						<IconButton aria-label="add" onClick={handleOpenAddit}>
							<AddIcon fontSize="small"></AddIcon>
						</IconButton>
					</Grid>
				</Grid>

				<Grid
					md={12}
					xs={12}
					columns={24}
					container
					item
					direction="row-reverse"
				>
					<Grid item md={3} xs={3} textAlign="center">
						<IconButton aria-label="settings">
							<SettingsIcon fontSize="small"></SettingsIcon>
						</IconButton>
					</Grid>
				</Grid>
			</Grid>

			<FeedAddit
				open={additOpen}
				onClose={handleCloseAddit}
				onSubmit={handleSubmitAddit}
			></FeedAddit>
		</div>
	);
}
