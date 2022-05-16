import FeedSidebar from "./feed-sidebar/index";
import NavBar from "./nav-bar/index";
import FeedItems from "./feed-items/index";
import { Pane } from "evergreen-ui";
import { NavbarContainer } from "./app.styled";

export function App() {
	return (
		<>
			<NavbarContainer>
				<NavBar></NavBar>
			</NavbarContainer>

			<Pane display="flex" flexDirection={"row"} padding={8}>
				<Pane flex={1} flexGrow={3} minWidth={320} >
					<FeedSidebar></FeedSidebar>
				</Pane>

				<Pane flex={1} flexGrow={4} minWidth={450}>
					<FeedItems></FeedItems>
				</Pane>

				<Pane flex={1} flexGrow={7}>这里是内容栏</Pane>
			</Pane>
		</>
	);
}
