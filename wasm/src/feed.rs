use wasm_bindgen::prelude::*;
use serde::{Deserialize};
use quick_xml::de::from_str;

use crate::http::*;

#[wasm_bindgen]
#[derive(Debug, Deserialize, PartialEq)]
#[serde(rename = "rss")]
pub struct RSSFeed {
    version: Option<String>
}

#[wasm_bindgen]
pub async fn get_feed_meta(url: String) -> RSSFeed {
    let body = http_get(url).await;
    let rss = from_str::<RSSFeed>(&body).unwrap();
    rss
}