import styled from "@emotion/styled";

export const PostViewContainer = styled.div`
	padding: 15px;
	display: flex;
	flex-direction: column;
	width: 100%;
	height: 100%;
	word-break: break-all;
	overflow: hidden;
	padding-right: 10px;

	&:hover {
		overflow-y: scroll;
		padding-right: 5px;
	}
`;

export const PostViewTitle = styled.div`
	font-size: 24px;
	width: 85%;
	font-weight: 600;
`;

export const PostViewTime = styled.div`
	width: 100%;
	margin-top: 15px;
	margin-bottom: 15px;
	font-size: 14px;
	color: #696f8c;
`;

export const PostViewBody = styled.div`
	margin-top: 15px;
	width: 100%;
	word-break: break-all;
	padding-right: 50px;

	line-height: 30px;

	img {
		border: solid 4px #d8dae5;
		border-radius: 6px;
		max-width: 500px;
	}

`;
