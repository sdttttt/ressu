mod utils;

use wasm_bindgen::prelude::*;

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;

#[wasm_bindgen]
extern "C" {

	#[wasm_bindgen(js_namespace = console)]
	fn log(s: &str);

	fn alert(s: &str);
}

macro_rules! console_log  {
	($($t:tt)*) => (log(&format_args!($($t)*).to_string()))
}

#[wasm_bindgen(start)]
pub fn on_start() {
	console_log!("WASM starting ...");
	let window = web_sys::window().expect("should have a window in this context");
	window.document().expect("window should have a document");
}

#[wasm_bindgen]
pub fn hello() {
	console_log!("from WASM greet!");
}

#[wasm_bindgen]
pub async fn http_get(url: String) -> String {
	let res = reqwest::get(url.as_str()).await.expect("http get Error");
	console_log!("Status: {}", res.status());

	let body = res.text().await.expect("response to utf-8 text");
	console_log!("Body:\n\n{}", body);
	
	body
}
