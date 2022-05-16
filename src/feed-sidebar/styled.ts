import styled from "@emotion/styled";
import { SocialMediaIcon } from "evergreen-ui";

export const SidebarContainer = styled.div`
	padding: 10px;
	display: flex;
	flex-direction: column;
`;

export const SearchInputContainer = styled.div`
	padding-top: 0px;
	width: 100%;
	padding-bottom: 10px;
`;

export const ChannelItemsContainer = styled.div`
	padding-top: 10px;
	flex-direction: column;
`;

export const ChannelItem = styled.div`
	display: flex;
	flex: 1;
	flex-direction: row;
	width: 100%;
	padding-top: 8px;
	padding-left: 15px;
	padding-right: 30px;
	font-size: 13px;
	text-align: left;
	padding-bottom: 8px;
	border-radius: 6px;

	&:hover {
		color: #3366ff;
		background-color: #f3f6ff;
	}
`;

export const ChannelItemText = styled.div<{selected: boolean}>`
	font-size: 13px;
	transition: all .5s;
	${({ selected }) => selected ? "font-weight: bold;": "" }
`;

export const ChannelItemIcon = styled.div``;

export const ChannelItemLoading = styled.div`
	height: 100%;
	flex-shrink: 0;
	text-align: end;
	flex-grow: 1;
	padding-top: 1px;
`;

export const ChannelItemSelectedBar = styled.div<{ open: boolean }>`
	border-radius: 6px;
	background-color: #5C85FF;
	width: 6px;
	margin-right: 10px;
	transition: all .1s;
	opacity: ${({open}) => open ? 1 : 0};
`;

export const rotateNinjaIcon = styled(SocialMediaIcon)`
	animation:fadenum 5s linear infinite;

	@keyframes fadenum {
		100% {
			transform: rotate(-720deg);
		}
	}
`;
