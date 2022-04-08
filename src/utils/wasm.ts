import store from "@store/index";
import { initWASMAsync } from "@store/wasm"

/**
 * init wasm and run func.
 * @param func callback
 */
export const runWASM = async (func: () => void) => {
	await initWASM();
	func();
};

/**
 * generate with args wasm runtime func.
 */
export const runWASMWithArgs = <R = any>(func: (...args: unknown[]) => R) => {
	return async (...arge: unknown[]): Promise<R> => {
		await initWASM();
		return func(...arge);
	}
}

/**
 * init wasm.
 */
export const initWASM = async () => {
	const { wasm: { ready } } = store.getState();
	if (!ready) {
		await store.dispatch(initWASMAsync());
	}
}