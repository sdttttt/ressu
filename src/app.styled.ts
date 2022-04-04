import styled from "@emotion/styled"
import { bodyHeight, navbarHeight } from "@/global.style";

export const Main = styled.div`
	height: ${bodyHeight};
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
