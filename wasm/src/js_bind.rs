use wasm_bindgen::prelude::*;

#[wasm_bindgen]
extern "C" {

    #[wasm_bindgen(js_namespace = console)]
    pub fn log(s: &str);

    pub fn alert(s: &str);

    #[wasm_bindgen(module = "src/utils/http")]
    pub async fn fetchRSSText(url: &str) -> JsValue;
}
