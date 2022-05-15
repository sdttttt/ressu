import styled from "@emotion/styled";

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
	padding-top: 8px;
	padding-left: 15px;
	font-size: 16px;
	text-align: left;
	padding-bottom: 8px;
	border-radius: 6px;
	
	margin-bottom: 8px;

	&:hover {
		color: #3366FF;
		background-color: #F3F6FF;
	}
`;
