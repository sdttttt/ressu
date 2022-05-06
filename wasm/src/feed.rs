use fast_xml::{de::from_str, events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::FromXml;
use crate::buf::BufPool;
use crate::item::ChannelItem;
use crate::{constants::*, utils::attrs_get_str, utils::reader_get_text};


#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "rss")]
pub struct RSSChannel {
    version: Option<String>,

    title: Option<String>,

    url: Option<String>,

    posts: Vec<ChannelItem>,
}

impl FromXml for RSSChannel {

    fn from_xml<B: std::io::BufRead>(
        bufs: &BufPool,
        reader: &mut Reader<B>,
    ) -> fast_xml::Result<RSSChannel> {
        let mut version = None;
        let mut title = None;
        let mut url = None;
        let mut posts = Vec::<ChannelItem>::new();

        reader.trim_text(true);

        let mut buf = bufs.pop();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref e)) => match reader.decode(e.name()).unwrap() {
                    "rss" => version = attrs_get_str(&reader, e.attributes(), "version").unwrap(),

                    "title" => title = attrs_get_str(&reader, e.attributes(), "title").unwrap(),

                    "url" => url = attrs_get_str(&reader, e.attributes(), "url").unwrap(),

                    "item" => {
                        let mut item_buf = Vec::new();

                        let mut item = ChannelItem::new();

                        reader.check_end_names(false);

                        loop {
                            match reader.read_event(&mut item_buf) {
                                Ok(Event::Start(ref e)) => match reader.decode(e.name()).unwrap() {
                                    "title" => {
                                        let text = reader_get_text(reader, e.name());
                                        console_log!("title => {}", &text);
                                        item.set_title(text.as_str());
                                    }

                                    "pubDate" => {
                                        let text = reader_get_text(reader, e.name());
                                        console_log!("pub_date => {}", &text);
                                        item.set_pub_date(text.as_str());
                                    }

                                    "link" => {
                                        let text = reader_get_text(reader, e.name());
                                        console_log!("link => {}", &text);
                                        item.set_link(text.as_str());
                                    }

                                    "description" => {
                                        let mut descript_buf = bufs.pop();
                                        reader.read_to_end(e.name(), &mut descript_buf).unwrap();

                                        console_log!(
                                            " ========== {}",
                                            reader.decode(&descript_buf).unwrap()
                                        );

                                        descript_buf.clear();
                                    }

                                    _ => {}
                                },

                                Ok(Event::End(ref e)) => match reader.decode(e.name()).unwrap() {
                                    "item" => break,
                                    _ => {}
                                },

                                Ok(Event::Text(ref e)) => {
                                    console_log!("{}", String::from_utf8_lossy(e));
                                }

                                Ok(Event::Eof) => break,

                                Ok(_) => {}

                                Err(fast_xml::Error::EndEventMismatch {
                                    expected: _,
                                    found: _,
                                }) => {}

                                Err(e) => panic!(
                                    "Error at position {}: {:?}",
                                    reader.buffer_position(),
                                    e
                                ),
                            }
                        }

                        reader.check_end_names(true);

                        posts.push(item);
                    }

                    _ => {}
                },

                Ok(Event::Text(ref e)) => {
                    console_log!("{}", String::from_utf8_lossy(e));
                }

                Ok(Event::Eof) => break,

                Ok(_) => {}

                Err(fast_xml::Error::EndEventMismatch {
                    expected: _,
                    found: _,
                }) => {}

                Err(e) => panic!("Error at position {}: {:?}", reader.buffer_position(), e),
            }
        }

        buf.clear();

        Ok(Self {
            version,
            title,
            url,
            posts,
        })
    }
}

#[wasm_bindgen]
impl RSSChannel {

    pub fn from_str(text: &str) -> RSSChannel {
        let mut bufs =  BufPool::new(4, 512);
        let mut reader = fast_xml::Reader::from_str(text);
        
        Self::from_xml(&mut bufs, &mut reader).unwrap()
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

    #[wasm_bindgen(js_name = postsLen)]
    pub fn posts_len(&self) -> usize {
        self.posts.len()
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
