import { toaster } from "evergreen-ui";
import { sleep } from ".";

export const successDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toaster.success(message);
};

export const notifyDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toaster.notify(message);
};

export const warningDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toaster.warning(message);
};

export const dangerDelay = async (message: string, delay: number) => {
	await sleep(delay);
	toaster.danger(message);
};
