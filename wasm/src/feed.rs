use wasm_bindgen::prelude::*;
use serde::{Deserialize, Serialize};
use quick_xml::de::from_str;

use crate::constants::*;

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "rss")]
pub struct RSSChannel {
    version: Option<String>
}

#[wasm_bindgen]
impl RSSChannel {
    
    #[wasm_bindgen(getter = version)]
    pub fn version(&self) -> Option<String>  {
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
pub async fn get_feed_meta(rss_text: String) -> RSSChannel  {
    let rss = from_str::<RSSChannel>(&rss_text).unwrap();
    rss
}

#[wasm_bindgen(js_name = getFeedJSON)]
pub async fn get_feed_json(rss_text: String) -> JsValue {
    get_feed_meta(rss_text).await.json()
}
