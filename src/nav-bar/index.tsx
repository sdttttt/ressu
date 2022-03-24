import { h } from "preact";
import classes from "./index.module.scss";
import { Button, Grid, IconButton } from "@mui/material";
import SettingsIcon from "@mui/icons-material/Settings";
import AddIcon from "@mui/icons-material/Add";

export default function NavBar() {
	return (
		<div class={classes["navbar-container"]}>
			<Grid container item columns={24}>
				<Grid md={12} xs={12} columns={24} container item>
					<Grid md={3} xs={3}>
						<IconButton aria-label="add">
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
					<Grid md={3} xs={3}>
						<IconButton aria-label="settings">
							<SettingsIcon fontSize="small"></SettingsIcon>
						</IconButton>
					</Grid>
				</Grid>
			</Grid>
		</div>
	);
}
