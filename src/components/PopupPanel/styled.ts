import styled from "@emotion/styled";
import { normalBorder } from "@/global.style";
import { computeFixedLeftDistanceFromPercent } from "@/styles/window";

export const Frame = styled.div`
	width: 100%;
	height: 100%;
	padding: 10px;
`;

export const ContainerShadow = styled.div<{
	open: boolean;
}>`
	position: fixed;
	top: 0;
	bottom: 0;
	left: 0;
	right: 0;
	width: 100%;
	height: 100%;
	background-color: black;
	opacity: ${({ open }) => (open ? "0.6" : "0")};
	z-index: ${({ open }) => (open ? "5" : "-1")};
	transition: all .3s;
`;

export const Container = styled.div<{
	open: boolean;
	borderColor?: string;
	height?: number;
}>`
	position: fixed;
	margin: 0 auto;
	top: ${props => {
		const { open, height = 200 } = props;
		if (open) {
			return -10;
		} else {
			return -10 - Number(height);
		}
	}}px;
	padding: 10px;
	width: 35%;
	left: ${() => computeFixedLeftDistanceFromPercent(35)};
	min-width: 300px;
	height: ${props => props.height || 200}px;
	box-shadow: 0px 0px 0px #eee;
	border: 1px ${props => props.borderColor || normalBorder} solid;
	background-color: #fff;
	border-radius: 4px;
	margin: 0 auto;
	z-index: 999;
	transition: top 0.3s;
`;

export const ConfirmDiv = styled.div`
	position: absolute;
	bottom: 10px;
	width: 95%;
	height: 30px;
	display: flex;
	justify-content: space-between;
	flex-direction: row;
`;

export const HeaderDiv = styled.div`
	padding-top: 10px;
	padding-left: 10px;
`;

export const TitleH4 = styled.h4`
	font-weight: 350;
	margin-bottom: 10px;
`;

export const DescriptionSpan = styled.span`
	font-weight: 350;
	font-size: 11px;
	color: #888;
`;
