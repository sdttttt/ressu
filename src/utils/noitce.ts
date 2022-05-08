import { toast } from "react-toastify";
import { sleep } from ".";

export const successDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toast.success(message);
};

export const notifyDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toast.info(message);
};

export const warningDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toast.warning(message);
};

export const dangerDelay = async (message: string, delay: number) => {
	await sleep(delay);
	
	toast.error(message);
};
