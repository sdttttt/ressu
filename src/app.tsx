import FeedSidebar from "./feed-sidebar/index";
import FeedItems from "./feed-items/index";
import { Pane } from "evergreen-ui";
import { AppContainer } from "./app.styled";

export function App() {
	return (
		<AppContainer>
			<Pane display="flex" flexDirection={"row"} padding={8} height="100%" overflow="hidden">
				<Pane flex={1} flexGrow={3} minWidth={320} height="100%">
					<FeedSidebar></FeedSidebar>
				</Pane>

				<Pane flex={1} flexGrow={4} minWidth={450} height="100%">
					<FeedItems></FeedItems>
				</Pane>

				<Pane flex={1} flexGrow={7}  height="100%">这里是内容栏</Pane>
			</Pane>
		</AppContainer>
	);
}
