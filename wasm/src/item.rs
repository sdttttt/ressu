use std::str::FromStr;

use fast_xml::Result;
use fast_xml::{events::Event, Reader};
use serde::{Serialize};
use wasm_bindgen::prelude::*;

use crate::buf::BufPool;
use crate::utils::TextOrCData;
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
    /// `Self::from_xml_with_reader(bufs, &mut reader)`
	/// 
	/// The `Self` is a reference to the current type, which is `ChannelItem`
	/// 
	/// Arguments:
	/// 
	/// * `bufs`: &BufPool - this is a pool of buffers that the parser uses to store the data it reads.
	/// * `text`: &str - the XML string to parse
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
	/// > The function reads the XML events from the reader, and when it encounters a start tag, it calls
	/// the `from_xml_with_reader` function of the corresponding field
	/// 
	/// Arguments:
	/// 
	/// * `bufs`: &BufPool
	/// * `reader`: &mut Reader<B>
	/// 
	/// Returns:
	/// 
	/// a Result<Self>
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
                    "title" => title =  TextOrCData::from_xml_with_reader(bufs, reader)?,

                    "pubDate" => pub_date =  TextOrCData::from_xml_with_reader(bufs, reader)?,

                    "link" => link = TextOrCData::from_xml_with_reader(bufs, reader)?,

                    "description" => description =  TextOrCData::from_xml_with_reader(bufs, reader)?,

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
