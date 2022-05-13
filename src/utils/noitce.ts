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
