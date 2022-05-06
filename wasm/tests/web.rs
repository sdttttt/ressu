//! Test suite for the Web and headless browsers.

#![cfg(target_arch = "wasm32")]

extern crate wasm_bindgen_test;
use std::assert_eq;
use std::include_str;

use wasm::feed::get_feed_json;
use wasm::feed::get_feed_meta;
use wasm::feed::RSSChannel;
use wasm_bindgen::JsValue;
use wasm_bindgen_test::console_log;
use wasm_bindgen_test::*;

wasm_bindgen_test_configure!(run_in_browser);

#[wasm_bindgen_test]
fn pass() {
    assert_eq!(1 + 1, 2);
}

#[wasm_bindgen_test]
fn test_parse_rss_from_str() {
    let rss_feed = RSSChannel::from_str(include_str!("../data/rss2.0.xml"));
    console_log!("{:?}", rss_feed);
    assert_eq!(rss_feed.is_specification(), true);

    console_log!("JsValue: {:?}", JsValue::from_serde(&rss_feed).unwrap());
}
