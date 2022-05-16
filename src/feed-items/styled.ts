import styled from "@emotion/styled";

export const PostItemContainer = styled.div`
	display: flex;
	width: 100%;
	height: 100%;
	flex-direction: column;
	padding-right: 10px;
`;

export const PostItem = styled.div`
	width: 100%;
	padding-top: 15px;
	padding-left: 15px;
	padding-bottom: 10px;
	padding-right: 15px;

	border-radius: 6px;

	&:hover {
		color: #3366ff;
		background-color: #f3f6ff;
	}
`;

export const PostItemTitle = styled.div`
	width: 100%;
	font-weight: bold;
	width: 100%;
	font-size: 16px;
	margin-bottom: 8px;
`;

export const PostItemBody = styled.div`
	font-size: 10px;
	display: -webkit-box;
	-webkit-line-clamp: 3;
	-webkit-box-orient: vertical;
	word-break: break-all;
	overflow: hidden;
	max-height: 120px;
	margin-bottom: 6px;
`;

export const PostItemTime = styled.div`
	font-size: 8px;
	color: #8f95b2;
	width: 100%;
	text-align: end;
`;
