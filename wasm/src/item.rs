use std::str::FromStr;

use fast_xml::Result;
use fast_xml::{events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::buf::BufPool;
use crate::utils::reader_get_text;
use crate::FromXmlWithReader;
use crate::FromXmlWithStr;
use crate::SkipThisElement;

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq, Default)]
#[serde(rename = "item")]
pub struct ChannelItem {
    title: Option<String>,

    description: Option<String>,

    pub_date: Option<String>,

    link: Option<String>,
}

impl FromXmlWithStr for ChannelItem {
    fn from_xml_with_str(bufs: &BufPool, text: &str) -> fast_xml::Result<ChannelItem> {
        let mut reader = Reader::from_str(text);

        let mut title = None;
        let mut description = None;
        let mut pub_date = None;
        let mut link = None;

        reader.trim_text(true);
        let mut buf = bufs.pop();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref e)) => match reader.decode(e.name()).unwrap() {
                    "title" => title = Some(reader_get_text(&mut reader, bufs)?),

                    "pubDate" => pub_date = Some(reader_get_text(&mut reader, bufs)?),

                    "link" => link = Some(reader_get_text(&mut reader, bufs)?),

                    "description" => description = Some(reader_get_text(&mut reader, bufs)?),

                    _ => {
                        SkipThisElement::from_xml_with_reader(bufs, &mut reader)?;
                    }
                },

                Ok(Event::Eof) => break,

                Ok(_) => {}

                Err(fast_xml::Error::EndEventMismatch {
                    expected: _,
                    found: _,
                }) => {}

                Err(e) => return Err(e),
            }
            buf.clear();
        }

        Ok(Self {
            title,
            description,
            pub_date,
            link,
        })
    }
}

#[wasm_bindgen]
impl ChannelItem {
    #[wasm_bindgen]
    pub fn title(&self) -> String {
        self.title.as_deref().unwrap_or("").to_string()
    }

    #[wasm_bindgen]
    pub fn description(&self) -> String {
        self.description.as_deref().unwrap_or("").to_string()
    }

    #[wasm_bindgen]
    pub fn pub_date(&self) -> String {
        self.pub_date.as_deref().unwrap_or("").to_string()
    }

    #[wasm_bindgen]
    pub fn link(&self) -> String {
        self.link.as_deref().unwrap_or("").to_string()
    }
}

impl FromStr for ChannelItem {
    type Err = fast_xml::Error;

    fn from_str(text: &str) -> Result<ChannelItem> {
        let bufs = BufPool::new(64, 8192);

        Self::from_xml_with_str(&bufs, text)
    }
}

impl ChannelItem {
    pub fn set_title(&mut self, title: &str) {
        self.title = Some(title.to_string());
    }

    pub fn set_description(&mut self, description: &str) {
        self.description = Some(description.to_string());
    }

    pub fn set_pub_date(&mut self, pub_date: &str) {
        self.pub_date = Some(pub_date.to_string());
    }

    pub fn set_link(&mut self, link: &str) {
        self.link = Some(link.to_string());
    }
}
