import init from "wasm";

const wasmState = {
	ready: false
};

/**
 * run func in wasm runtime.
 */
export const runWASM = async <R = any>(func: () => R): Promise<R> => {
	await initWASM();
	return func();
};

/**
 * generate with args wasm runtime func.
 */
export const runWASMWithArgs = <R = any>(func: (...args: unknown[]) => R) => {
	return async (...arge: unknown[]): Promise<R> => {
		await initWASM();
		return func(...arge);
	};
};

/**
 * init wasm.
 */
export const initWASM = async () => {
	const { ready } = wasmState;
	if (!ready) {
		await init();
		wasmState.ready = true;
	}
};
