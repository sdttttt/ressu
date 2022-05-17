import styled from "@emotion/styled";
import { navbarHeight } from "@/global.style";
import { Resizable } from "re-resizable";

export const AppContainer = styled.div`
	position: fixed;
	top: 0;
	right: 0;
	bottom: 0;
	left: 0;
`;

export const FeedSidebarContainer = styled.div`
	height: 100%;
	width: 30%;
`;


export const NavbarContainer = styled.div`
	height: ${navbarHeight};
	min-height: 44px;
`;

export const SplitBorder = styled.div`
	height: 100%;
	width: 1px;
	margin-right: 5px;
	margin-left: 5px;
	background-color: #E6E8F0;
	transition: all .2s;
`;
