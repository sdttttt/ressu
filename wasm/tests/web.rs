//! Test suite for the Web and headless browsers.

#![cfg(target_arch = "wasm32")]

extern crate wasm_bindgen_test;
use js_sys::Date;
use std::assert_eq;
use std::include_str;
use std::str::FromStr;
use wasm::feed::RSSChannel;
use wasm_bindgen_test::console_log;
use wasm_bindgen_test::*;
use web_sys;
use web_sys::Window;

wasm_bindgen_test_configure!(run_in_browser);

fn perf_now() -> f64 {
    let window: Window = web_sys::window().expect("should have a window in this context");

    let performance = window
        .performance()
        .expect("performance should be available");

    performance.now()
}

fn record_parse_rss_perf_from_text(rss_text: &str) {
	let start_time = perf_now();
	let rss = RSSChannel::from_str(rss_text).unwrap();
	console_log!("{:?}", rss);
  let end_time = perf_now();
  console_log!("Time: {}ms", end_time - start_time);
}

#[wasm_bindgen_test]
fn pass() {
    assert_eq!(1 + 1, 2);
}

#[wasm_bindgen_test]
fn test_parse_zdfx_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/ZDFX_RSS2.0.xml"));
	console_log!("{}", "ZDFS");
}

#[wasm_bindgen_test]
fn test_parse_3dm_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/3DM_RSS2.0.xml"));
	console_log!("{}", "3DM");
}

#[wasm_bindgen_test]
fn test_parse_genshin_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/GENSHIN_RSS2.0.xml"));
	console_log!("{}", "GENSHIN");
}

#[wasm_bindgen_test]
fn test_parse_els_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/ELS_RSS2.0.xml"));
	console_log!("{}", "ELS");
}

#[wasm_bindgen_test]
fn test_parse_ns_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/JUMP_PS5_ALL_RSS2.0.xml"));
	console_log!("{}", "JUMP_PS5_ALL");
}

#[wasm_bindgen_test]
fn test_parse_qidian_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/QIDIAN_RSS2.0.xml"));
	console_log!("{}", "GENSHIN");
}

#[wasm_bindgen_test]
fn test_parse_ps5_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/JUMP_NS_ALL_RSS2.0.xml"));
	console_log!("{}", "JUMP_NS_ALL");
}

#[wasm_bindgen_test]
fn test_parse_12306_rss_from_str() {

	record_parse_rss_perf_from_text(include_str!("../data/12306_RSS2.0.xml"));
	console_log!("{}", "12306");
}

#[wasm_bindgen_test]
fn test_parse_yystv_rss_from_str() {
	record_parse_rss_perf_from_text(include_str!("../data/YYSTV_RSS2.0.xml"));
	console_log!("{}", "YYSTV");
}