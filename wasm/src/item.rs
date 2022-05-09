use std::str::FromStr;

use fast_xml::Result;
use fast_xml::{events::Event, Reader};
use serde::{Serialize};
use wasm_bindgen::prelude::*;

use crate::buf::BufPool;
use crate::utils::reader_get_text;
use crate::FromXmlWithReader;
use crate::FromXmlWithStr;
use crate::SkipThisElement;

#[wasm_bindgen]
#[derive(Debug, Serialize, Default)]
#[serde(rename = "item")]
pub struct ChannelItem {
	
    title: Option<String>,

	
    description: Option<String>,

	#[serde(rename = "pubDate")]
    pub_date: Option<String>,

	
    link: Option<String>,
}

impl FromXmlWithStr for ChannelItem {
	
   /// It reads the XML and populates the fields of the struct
   /// 
   /// Arguments:
   /// 
   /// * `bufs`: &BufPool - this is a pool of buffers that are used to read the XML.
   /// * `text`: The XML text to parse.
   /// 
   /// Returns:
   /// 
   /// A ChannelItem
    fn from_xml_with_str(bufs: &BufPool, text: &str) -> fast_xml::Result<ChannelItem> {
        let mut reader = Reader::from_str(text);

        Self::from_xml_with_reader(bufs, &mut reader)
    }
}


impl FromXmlWithReader for ChannelItem {
	fn from_xml_with_reader<B: std::io::BufRead>(bufs: &BufPool, reader: &mut Reader<B>) -> fast_xml::Result<Self> {
		
        let mut title = None;
        let mut description = None;
        let mut pub_date = None;
        let mut link = None;

        reader.trim_text(true);
        let mut buf = bufs.pop();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref e)) => match reader.decode(e.name())? {
                    "title" => title = Some(reader_get_text(reader, bufs)?),

                    "pubDate" => pub_date = Some(reader_get_text(reader, bufs)?),

                    "link" => link = Some(reader_get_text(reader, bufs)?),

                    "description" => description = Some(reader_get_text(reader, bufs)?),

                    _ => {
                        SkipThisElement::from_xml_with_reader(bufs, reader)?;
                    }
                },

                Ok(Event::Eof | Event::End(_)) => break,

                Ok(_) => (),

                Err(e) => {
					return Err(e) 
				},
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
    #[wasm_bindgen(getter)]
    pub fn title(&self) -> String {
        self.title.as_deref().unwrap_or("").to_string()
    }

    #[wasm_bindgen(getter)]
    pub fn description(&self) -> String {
        self.description.as_deref().unwrap_or("").to_string()
    }

    #[wasm_bindgen(getter, js_name = pubDate)]
    pub fn pub_date(&self) -> String {
        self.pub_date.as_deref().unwrap_or("").to_string()
    }

    #[wasm_bindgen(getter)]
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
