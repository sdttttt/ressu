import initWASM, { http_get  } from "wasm";
import { WASMActions } from "@store/actions";
import store from "@store/index";

export const runWASM = async (func: () => void) => {
	const { hasInitWASM } = store.getState();
	if (!hasInitWASM) {
		console.log("JS: init WASM...");
		await initWASM();
		store.dispatch(WASMActions.initializeWASM());
	}
	func();
};
