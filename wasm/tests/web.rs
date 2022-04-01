//! Test suite for the Web and headless browsers.

#![cfg(target_arch = "wasm32")]

extern crate wasm_bindgen_test;
use wasm::greet;
use wasm_bindgen_test::{console_log};
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);


#[wasm_bindgen_test]
fn pass() {
    assert_eq!(1 + 1, 2);
}

#[wasm_bindgen_test]
async fn simple_example() {
    let res = reqwest::get("https://hyper.rs")
        .await
        .expect("http get example");
    console_log!("Status: {}", res.status());
	
    let body = res.text().await.expect("response to utf-8 text");
    console_log!("Body:\n\n{}", body);
}