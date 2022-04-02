import styled from "styled-components";
import { normalBorder } from "@/global.style";
import { computeFixedLeftDistanceFromPercent } from "@/styles/window";

export const Frame = styled.div`
	height: 100%;
	width: 100%;
	pointer-events: none;
	z-index: -1;
`;


export const Container = styled.div<{
	open: boolean,
	borderColor?: string,
	height?: number,
}>`
	position: fixed;
	margin: 0 auto;
	top: ${
		props => {
			const { open, height = 200 } = props;
			if (open) {
				return -10;
			} else {
				return (-10 - Number(height));
			}
		}
	}px;
	padding: 10px;
	width: 35%;
	left: ${() => computeFixedLeftDistanceFromPercent(35)};
	min-width: 300px;
	height: ${props => props.height || 200}px; 
	box-shadow: 1px 1px 10px #EEE;
	border:  1px ${props => props.borderColor || normalBorder} solid;
	background-color: #FFF;
	border-radius: 4px;
	margin: 0  auto;
	z-index: 999;
	transition: top .5s;
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
