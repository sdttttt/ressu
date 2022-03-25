import FeedSidebar from "./feed-sidebar/index";
import NavBar from "./nav-bar/index";
import FeedItems from "./feed-items/index";
import classes from "./app.module.scss";
import { Grid } from "@mui/material";

export function App() {
	
	return (
		<>
			<div className={classes.navbar}>
				<NavBar></NavBar>
			</div>
			<Grid container className={classes.main} columns={24}>
				<Grid item  md={6} xs={8} alignContent="center">
					<FeedSidebar></FeedSidebar>
				</Grid>

				<Grid item md={6} xs={8} alignContent="center">
					<FeedItems></FeedItems>
				</Grid>

				<Grid item md={24 - 6 - 6} alignContent="center">
					这里是内容栏
				</Grid>
			</Grid>
		</>
	);
}
