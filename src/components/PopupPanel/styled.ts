import styled from "styled-components";
import { normalBorder } from "@/global.style";
import { computeFixedLeftDistanceFromPercent } from "@/styles/window";

export const Frame = styled.div`
	height: 100%;
	width: 100%;
	pointer-events: none;
	z-index: -1;
`;

type ContainerPropsType = {
	borderColor?: string,
	height?: string,
}

export const Container = styled.div<ContainerPropsType>`
	position: fixed;
	margin: 0 auto;
	top: -10px;
	padding: 10px;
	width: 35%;
	left: ${() => computeFixedLeftDistanceFromPercent(35)};
	min-width: 300px;
	height: ${props => props.height || "200px"};
	box-shadow: 1px 1px 10px #EEE;
	border:  1px ${props => props.borderColor || normalBorder} solid;
	background-color: #FFF;
	border-radius: 4px;
	margin: 0  auto;
	z-index: 999;
`;
