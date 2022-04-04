import FeedSidebar from "./feed-sidebar/index";
import NavBar from "./nav-bar/index";
import FeedItems from "./feed-items/index";
import { Pane } from "evergreen-ui";
import { NavbarContainer  } from "./app.styled"

export function App() {

	return (
		<>
			<NavbarContainer>
				<NavBar></NavBar>
			</NavbarContainer>

			<Pane display="flex" padding={8}>
				<Pane flex={1}>
				<FeedSidebar></FeedSidebar>
				</Pane>

				<Pane  flex={1}>
				<FeedItems></FeedItems>
				</Pane>

				<Pane  flex={2}>
				这里是内容栏
				</Pane>
			</Pane>
		</>
	);
}
