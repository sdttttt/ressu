import FeedSidebar from "./feed-sidebar/index";
import NavBar from "./nav-bar/index";
import FeedItems from "./feed-items/index";
import classes from "./app.module.scss";
import { Grid } from "@mui/material";

export function App() {
	return (
		<>
			<div class={classes.navbar}>
				<NavBar></NavBar>
			</div>
			<Grid container item class={classes.main} columns={24}>
				<Grid md={6} xs={8} alignContent="center">
					<FeedSidebar></FeedSidebar>
				</Grid>

				<Grid md={6} xs={8} alignContent="center">
					<FeedItems></FeedItems>
				</Grid>

				<Grid md={24 - 6 - 6} xs={24 - 8 - 8} alignContent="center">
					这里是内容栏
				</Grid>
			</Grid>
		</>
	);
}
