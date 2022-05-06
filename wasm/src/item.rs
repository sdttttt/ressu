
use std::borrow::BorrowMut;

use fast_xml::{events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::FromXml;
use crate::buf::BufPool;
use crate::utils::reader_get_text;

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "item")]
pub struct ChannelItem {
    title: Option<String>,

    description: Option<String>,

    pub_date: Option<String>,

    link: Option<String>,
}

impl FromXml  for ChannelItem {
    fn from_xml(bufs: &BufPool, text: &str) -> fast_xml::Result<ChannelItem> {

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
                    "title" => {
                        let mut title_buf = bufs.pop();
                        title = Some(
                            reader_get_text(&mut reader, e.name(), &mut title_buf)
                        )
                    }

                    "pubDate" => {
                        let mut pub_date_buf = bufs.pop();
                        pub_date = Some(
                            reader_get_text(&mut reader, e.name(), &mut pub_date_buf)
                        )
                    }

                    "link" => {
                        let mut link_buf = bufs.pop();
                        link = Some(
                            reader_get_text(&mut reader, e.name(), &mut link_buf)
                        )
                    }

                    "description" => {
                        let mut description_buf = bufs.pop();

                        let start_position = reader.buffer_position();

                        reader.check_end_names(false);
                        loop {
                            match reader.read_event(&mut description_buf) {
                                Ok(Event::End(ref e)) => match reader.decode(e.name()).unwrap() {
                                    "description" => break,
                                    _ => {}
                                },
                                Ok(Event::Eof) => break,
                                _ => {}
                            }
                        }
                        reader.check_end_names(true);
                        
                        let end_position = reader.buffer_position();

                        let end_tag_len = "</description>".len();

                        let text_string = text.to_string();
                        let item_slice = &text_string.as_bytes()[start_position..end_position - end_tag_len];
                    
                        description = Some(String::from_utf8_lossy(item_slice).borrow_mut().to_string());
                    }

                    _ => {}
                },

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
		self.title.as_deref().unwrap_or_else(|| "").to_string()
	}


	#[wasm_bindgen]
	pub fn description(&self) -> String {
		self.description.as_deref().unwrap_or_else(|| "").to_string()
	}



	#[wasm_bindgen]
	pub fn pub_date(&self) -> String {
		self.pub_date.as_deref().unwrap_or_else(|| "").to_string()
	}

	#[wasm_bindgen]
	pub fn link(&self) -> String {
		self.link.as_deref().unwrap_or_else(|| "").to_string()
	}

}

impl ChannelItem {

    pub fn from_str(text: &str) -> ChannelItem {
        let mut bufs = BufPool::new(16, 512);
        
        Self::from_xml(&mut bufs, text).unwrap()
    }


	pub fn new() -> Self {
		Self {
			title: None,
			description: None,
			pub_date: None,
			link: None,
		}
	}

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