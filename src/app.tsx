import FeedSidebar from "./feed-sidebar/index";
import FeedItems from "./feed-items/index";
import { Pane } from "evergreen-ui";
import { AppContainer, SidebarSplit } from "./app.styled";
import PostView from "@/post-view/index";

export function App() {
	return (
		<AppContainer>
			<Pane display="flex" flexDirection={"row"} height="100%" overflow="hidden">
				<Pane flex={1} flexGrow={3} minWidth={320} height="100%">
					<FeedSidebar></FeedSidebar>
				</Pane>
				
				<SidebarSplit></SidebarSplit>

				<Pane flex={1} flexGrow={4} minWidth={450} height="100%">
					<FeedItems></FeedItems>
				</Pane>

				<Pane flex={1} flexGrow={7}  height="100%">
					<PostView></PostView>
				</Pane>
			</Pane>
		</AppContainer>
	);
}
