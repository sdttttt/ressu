import FeedSidebar from "./feed-sidebar/index";
import FeedItems from "./feed-items/index";
import { Pane } from "evergreen-ui";
import { AppContainer, SplitBorder } from "./app.styled";
import PostView from "@/post-view/index";
import { Resizable } from "re-resizable"

export function App() {
	return (
		<AppContainer>
			<Pane display="flex" flexDirection={"row"} height="100%" overflow="hidden">
				<Resizable defaultSize={{ height: "100%", width: "20%" }}>
					<FeedSidebar></FeedSidebar>
				</Resizable>

				<SplitBorder></SplitBorder>

				<Resizable defaultSize={{ height: "100%", width: "25%" }}>
					<FeedItems></FeedItems>
				</Resizable>

				<SplitBorder></SplitBorder>

				<Pane flex={1} flexGrow={7}  height="100%">
					<PostView></PostView>
				</Pane>
			</Pane>
		</AppContainer>
	);
}
