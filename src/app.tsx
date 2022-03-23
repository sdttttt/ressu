import FeedSidebar from "./feed-sidebar/index";
import NavBar from "./nav-bar/index";
import classes from "./app.module.scss";

export function App() {
	return (
		<>
			<div class={classes.navbar}>
				<NavBar></NavBar>
			</div>
			<div class={classes["feed-sidebar"]}>
				<FeedSidebar></FeedSidebar>
			</div>

			<div>
				Content
			</div>
		</>
	);
}
