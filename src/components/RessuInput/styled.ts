import { normalBorder } from "@/global.style"
import styled from "@emotion/styled";

export const InputContainer = styled.div`
	background-color: ${normalBorder};
	height: 33px;
	border-radius: 3px;
	display: flex;
	flex-direction: row;
	padding: 5px;
`;

export  const InputPrefix = styled.div`
	padding-top: 3px;
	padding-left: 3px;
`;

export const InInput = styled.input`
	background-color: ${normalBorder};
	color: rgb(88, 88, 88);
	height: 90%;
	width: 99%;
	border: 0;
	font-size: 14px;
	outline: none;
	padding-left: 5px;
	padding-top: 2px;
`;
