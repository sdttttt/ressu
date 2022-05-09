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

wasm_bindgen_test_configure!(run_in_browser);

#[wasm_bindgen_test]
fn pass() {
    assert_eq!(1 + 1, 2);
}

#[wasm_bindgen_test]
fn test_parse_zdfx_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/ZDFX_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);
    
	assert_eq!(rss_feed.is_specification(), true);


}

#[wasm_bindgen_test]
fn test_parse_3dm_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/3DM_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);

	assert_eq!(rss_feed.is_specification(), true);
}

#[wasm_bindgen_test]
fn test_parse_genshin_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/GENSHIN_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);

    assert_eq!(rss_feed.is_specification(), true);

}

#[wasm_bindgen_test]
fn test_parse_els_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/ELS_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);

    assert_eq!(rss_feed.is_specification(), true);
}

#[wasm_bindgen_test]
fn test_parse_ns_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/JUMP_PS5_ALL_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);

    assert_eq!(rss_feed.is_specification(), true);
}

#[wasm_bindgen_test]
fn test_parse_qidian_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/QIDIAN_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);

    assert_eq!(rss_feed.is_specification(), true);
}

#[wasm_bindgen_test]
fn test_parse_ps5_rss_from_str() {
    let start_time = Date::now();

    let rss_feed = RSSChannel::from_str(include_str!("../data/JUMP_NS_ALL_RSS2.0.xml")).unwrap();
    let end_time = Date::now();
    console_log!("Time: {}ms", end_time - start_time);

    assert_eq!(rss_feed.is_specification(), true);
}
