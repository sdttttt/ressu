use wasm_bindgen::prelude::*;

#[wasm_bindgen]
extern "C" {

    #[wasm_bindgen(js_namespace = console)]
    pub fn log(s: &str);

    pub fn alert(s: &str);
}


#[wasm_bindgen(module = "@tauri-apps/api/http")]
extern  "C" {
	pub async fn fetch(url: &str);
}
