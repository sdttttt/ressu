use quick_xml::de::from_str;
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::{constants::*, js_bind::fetchRSSText};

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "rss")]
pub struct RSSChannel {
    version: Option<String>,
}

#[wasm_bindgen]
impl RSSChannel {
    #[wasm_bindgen(getter = version)]
    pub fn version(&self) -> Option<String> {
        self.version.clone()
    }

    #[wasm_bindgen(js_name = isSpecification)]
    pub fn is_specification(&self) -> bool {
        self.version.as_deref().unwrap_or("") == RSS_VERSION_AVAILABLE
    }

    #[wasm_bindgen]
    pub fn json(&self) -> JsValue {
        JsValue::from_serde(self).unwrap()
    }
}

#[wasm_bindgen(js_name = getFeedMeta)]
pub fn get_feed_meta(rss_text: &str) -> RSSChannel {
    let rss = from_str::<RSSChannel>(&rss_text).unwrap();
    rss
}

#[wasm_bindgen(js_name = getFeedJSON)]
pub fn get_feed_json(rss_text: &str) -> JsValue {
    get_feed_meta(rss_text).json()
}

#[wasm_bindgen(js_name = getFeedMetaFormUrl)]
pub async fn get_feed_meta_from_url(url: String) -> RSSChannel {
    let text = fetchRSSText(url.as_str()).await.as_string().unwrap();
    get_feed_meta(text.as_str())
}

#[wasm_bindgen(js_name = getFeedJSONFormUrl)]
pub async fn get_feed_json_form_url(url: String) -> JsValue {
    let text = fetchRSSText(url.as_str()).await.as_string().unwrap();
    get_feed_json(text.as_str())
}