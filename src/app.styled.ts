import styled from "@emotion/styled";
import { navbarHeight } from "@/global.style";
import { Pane } from "evergreen-ui";

export const AppContainer = styled.div`
	position: fixed;
	top: 0;
	right: 0;
	bottom: 0;
	left: 0;
`;

export const FeedSidebarContainer = styled.div`
	height: 100%;
	width: 300px;
	top: ${navbarHeight};
`;

export const NavbarContainer = styled.div`
	height: ${navbarHeight};
	min-height: 44px;
`;

export const SidebarSplit = styled.div`
	height: 100%;
	width: 1px;
	margin-right: 5px;
	background-color: #E6E8F0;
	transition: all .2s;
`;
