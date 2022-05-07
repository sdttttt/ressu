use std::borrow::BorrowMut;

use fast_xml::{de::from_str, events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::buf::BufPool;
use crate::item::ChannelItem;
use crate::utils::{reader_get_sub_node_str, reader_get_text};
use crate::{constants::*, utils::attrs_get_str};
use crate::{FromXmlWithReader, FromXmlWithStr, SkipThisElement};

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "rss")]
pub struct RSSChannel {
    version: Option<String>,

    title: Option<String>,

    url: Option<String>,

    posts: Vec<ChannelItem>,
}

impl FromXmlWithStr for RSSChannel {
    fn from_xml_with_str(bufs: &BufPool, text: &str) -> fast_xml::Result<RSSChannel> {
        let mut reader = Reader::from_str(text);

        let mut version = None;
        let mut title = None;
        let mut url = None;
        let mut posts = Vec::<ChannelItem>::new();

        reader.trim_text(true);

        let mut buf = bufs.pop();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref re)) => match reader.decode(re.name())? {
                    "rss" => version = attrs_get_str(&reader, re.attributes(), "version")?,

                    "channel" => {
                        let mut cbuf = bufs.pop();

                        let mut count = 0u64;

                        loop {
                            match reader.read_event(&mut cbuf) {
                                Ok(Event::Start(ref ce)) => match reader.decode(ce.name())? {
                                    "title" => title = Some(reader_get_text(&mut reader, bufs)?),

                                    "link" => url = Some(reader_get_text(&mut reader, bufs)?),

                                    "item" => {
                                        let tag = reader.decode(ce.name())?;
                                        let item_slice =
                                            reader_get_sub_node_str(&mut reader, bufs, tag, text)?;

                                        let item = ChannelItem::from_str(&item_slice);
                                        posts.push(item);
                                        count += 1;
                                    }

                                    _ => {
                                        SkipThisElement::from_xml_with_reader(bufs, &mut reader)?;
                                    }
                                },
                                Ok(Event::End(ref ce)) => match reader.decode(ce.name())? {
                                    "channel" => break,
                                    _ => (),
                                },
                                Ok(Event::Eof) => break,
                                _ => (),
                            }
                            cbuf.clear()
                        }

                        console_log!("item count: {}", count);
                    }

                    _ => (),
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
            buf.clear();
        }

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

impl RSSChannel {
    pub fn from_str(text: &str) -> RSSChannel {
        let mut bufs = BufPool::new(4, 2048);

        Self::from_xml_with_str(&mut bufs, text).unwrap()
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
