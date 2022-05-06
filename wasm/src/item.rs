
use fast_xml::{events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

#[wasm_bindgen]
#[derive(Debug, Deserialize, Serialize, PartialEq)]
#[serde(rename = "item")]
pub struct ChannelItem {
    title: Option<String>,

    description: Option<String>,

    pub_date: Option<String>,

    link: Option<String>,
}

#[wasm_bindgen]
impl ChannelItem {

    pub fn from_str(text: &str) -> Self {
        let mut title = None;
        let mut description = None;
        let mut pub_date = None;
        let mut link = None;

        let mut reader = Reader::from_str(&text);
        reader.trim_text(true);
        let mut buf = Vec::new();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref e)) => match e.name() {
                    b"title" => {
                        title = Some(
                            reader
                                .read_text(e.name(), &mut Vec::new())
                                .unwrap()
                                .to_string(),
                        )
                    }

                    b"description" => {
                        description = Some(
                            reader
                                .read_text(e.name(), &mut Vec::new())
                                .unwrap()
                                .to_string(),
                        )
                    }

                    b"pubDate" => {
                        pub_date = Some(
                            reader
                                .read_text(e.name(), &mut Vec::new())
                                .unwrap()
                                .to_string(),
                        )
                    }

                    b"link" => {
                        link = Some(
                            reader
                                .read_text(e.name(), &mut Vec::new())
                                .unwrap()
                                .to_string(),
                        )
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

        Self {
            title,
            description,
            pub_date,
            link,
        }
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

	#[wasm_bindgen]
	pub fn title(&self) -> String {
		self.title.as_deref().unwrap_or_else(|| "").to_string()
	}

	pub fn set_description(&mut self, description: &str) {
		self.description = Some(description.to_string());
	}

	#[wasm_bindgen]
	pub fn description(&self) -> String {
		self.description.as_deref().unwrap_or_else(|| "").to_string()
	}

	pub fn set_pub_date(&mut self, pub_date: &str) {
		self.pub_date = Some(pub_date.to_string());
	}

	#[wasm_bindgen]
	pub fn pub_date(&self) -> String {
		self.pub_date.as_deref().unwrap_or_else(|| "").to_string()
	}

	pub fn set_link(&mut self, link: &str) {
		self.link = Some(link.to_string());
	}

	#[wasm_bindgen]
	pub fn link(&self) -> String {
		self.link.as_deref().unwrap_or_else(|| "").to_string()
	}

}
