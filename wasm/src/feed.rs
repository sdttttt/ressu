use quick_xml::{de::from_str, events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::{constants::*, utils::attrs_get_str};

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "rss")]
pub struct RSSChannel {
    version: Option<String>,

    title: Option<String>,

    url: Option<String>,
}

#[wasm_bindgen]
impl RSSChannel {
    pub fn from_str(text: &str) -> Self {
        let mut version = None;
        let mut title = None;
        let mut url = None;

        let mut reader = Reader::from_str(&text);
        reader.trim_text(true);
        let mut buf = Vec::new();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref e)) => match e.name() {
                    b"rss" => {
                        version = attrs_get_str(&mut reader, e.attributes(), "version").unwrap();
                    }

                    b"title" => {
                        title = attrs_get_str(&mut reader, e.attributes(), "title").unwrap();
                    }

                    b"url" => {
                        url = attrs_get_str(&mut reader, e.attributes(), "url").unwrap();
                    }

                    _ => {}
                },

                Ok(Event::Eof) => break,

                Ok(_) => {}

                Err(quick_xml::Error::EndEventMismatch { expected: _, found: _ }) => {}

                Err(e) => panic!("Error at position {}: {:?}", reader.buffer_position(), e),
            }
        }

        Self {
            version,
            title,
            url,
        }
    }

    #[wasm_bindgen(getter = version)]
    pub fn version(&self) -> String {
        self.version.as_deref().unwrap_or("").to_string()
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
