import { toast } from "react-hot-toast";
import { sleep } from ".";

export const successDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toast.success(message);
};

export const dangerDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toast.error(message);
};

export const infoDelay = async (message: string, delay = 0) => {
	await sleep(delay);

	toast.success(message, {
		style: {
			border: "1px solid #713200",
			padding: "16px",
			color: "#713200"
		},
		iconTheme: {
			primary: "#713200",
			secondary: "#FFFAEE"
		}
	});
};
