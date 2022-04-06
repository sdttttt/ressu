import store from "@store/index";
import { initWASMAsync } from "@store/wasm"

export const runWASM = async (func: () => void) => {
	const { wasm: {ready} } = store.getState();
	if (!ready) {
		await store.dispatch(initWASMAsync());
	}
	func();
};
