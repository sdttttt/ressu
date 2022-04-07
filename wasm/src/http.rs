use wasm_bindgen::prelude::*;

#[wasm_bindgen]
pub async fn http_get(url: String) -> String {
    let res = reqwest::get(url.as_str()).await.expect("http get Error");
    console_log!("Status: {}", res.status());

    let body = res.text().await.expect("response to utf-8 text");
    console_log!("Body:\n\n{}", body);

    body
}