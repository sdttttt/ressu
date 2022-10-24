use std::str::FromStr;

use fast_xml::{events::Event, Reader};
use serde::{Deserialize, Serialize};
use wasm_bindgen::prelude::*;

use crate::buf::BufPool;
use crate::item::ChannelItem;
use crate::utils::{parse_atom_link, AtomLink, NumberData, TextOrCData};
use crate::{constants::*, utils::attrs_get_str};
use crate::{FromXmlWithReader, FromXmlWithStr, SkipThisElement};

#[wasm_bindgen]
#[derive(Debug, Serialize, Deserialize)]
#[serde(rename = "rss")]
pub struct RSSChannel {
    #[serde(skip)]
    version: Option<String>,

    title: Option<String>,

    description: Option<String>,

    url: Option<String>,

    #[serde(rename = "atomLink")]
    atom_link: Option<String>,

    language: Option<String>,

    #[serde(rename = "webMaster")]
    web_master: Option<String>,

    generator: Option<String>,

    #[serde(rename = "lastBuildDate")]
    last_build_date: Option<String>,

    ttl: Option<usize>,

    image: Option<ChannelImage>,

    posts: Vec<ChannelItem>,

    copyright: Option<String>,
}

impl FromXmlWithStr for RSSChannel {
    /// > It takes a string, creates a reader from it, and then calls the function that takes a reader
    ///
    /// Arguments:
    ///
    /// * `bufs`: &BufPool - this is a pool of buffers that the parser uses to store the data it reads.
    /// * `text`: The XML text to parse.
    ///
    /// Returns:
    ///
    /// A `fast_xml::Result<RSSChannel>`
    fn from_xml_with_str(bufs: &BufPool, text: &str) -> fast_xml::Result<RSSChannel> {
        let mut reader = Reader::from_str(text);

        Self::from_xml_with_reader(bufs, &mut reader)
    }
}

impl FromXmlWithReader for RSSChannel {
    /// > Read the XML document, and for each element, if it's a `<rss>` element, get the `version`
    /// attribute, if it's a `<channel>` element, read the `<title>`, `<description>`, `<link>` and
    /// `<item>` elements, and if it's a `<item>` element, read the `<title>`, `<description>`, `<link>`,
    /// `<pubDate>`, `<guid>` and `<category>` elements
    ///
    /// Arguments:
    ///
    /// * `bufs`: &BufPool - this is a pool of buffers that are used to read the XML.
    /// * `reader`: &mut Reader<B>
    ///
    /// Returns:
    ///
    /// A `Result` of `Channel`
    fn from_xml_with_reader<B: std::io::BufRead>(
        bufs: &BufPool,
        reader: &mut Reader<B>,
    ) -> fast_xml::Result<Self> {
        let mut version = None;
        let mut title = None;
        let mut description = None;
        let mut url = None;
        let mut atom_link = None;
        let mut language = None;
        let mut web_master = None;
        let mut last_build_date = None;
        let mut ttl = None;
        let mut image = None;
        let mut posts = Vec::<ChannelItem>::new();
        let mut copyright = None;
        let mut generator = None;

        reader.trim_text(true);

        let mut buf = bufs.pop();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref re)) => match reader.decode(re.local_name())? {
                    "rss" => version = attrs_get_str(&reader, re.attributes(), "version")?,

                    "channel" => {
                        let mut cbuf = bufs.pop();

                        loop {
                            match reader.read_event(&mut cbuf) {
                                Ok(Event::Empty(ref ce)) => match reader.decode(ce.local_name())? {
                                    "link" => match parse_atom_link(reader, ce.attributes())? {
                                        Some(AtomLink::Alternate(link)) => url = Some(link),
                                        Some(AtomLink::Source(link)) => atom_link = Some(link),
                                        _ => {}
                                    },

                                    _ => (),
                                },

                                Ok(Event::Start(ref ce)) => match reader.decode(ce.local_name())? {
                                    "title" => {
                                        title = TextOrCData::from_xml_with_reader(bufs, reader)?
                                    }

                                    "description" => {
                                        description =
                                            TextOrCData::from_xml_with_reader(bufs, reader)?;
                                    }

                                    "link" => {
                                        match TextOrCData::from_xml_with_reader(bufs, reader)? {
                                            Some(s) => url = Some(s),
                                            None => {
                                                match parse_atom_link(reader, ce.attributes())? {
                                                    Some(AtomLink::Source(e)) => {
                                                        atom_link = Some(e)
                                                    }
                                                    Some(AtomLink::Alternate(e)) => url = Some(e),
                                                    _ => (),
                                                }
                                            }
                                        };
                                    }

                                    "language" => {
                                        language = TextOrCData::from_xml_with_reader(bufs, reader)?
                                    }

                                    "webMaster" => {
                                        web_master =
                                            TextOrCData::from_xml_with_reader(bufs, reader)?
                                    }

                                    "generator" => {
                                        generator = TextOrCData::from_xml_with_reader(bufs, reader)?
                                    }

                                    "lastBuildDate" => {
                                        last_build_date =
                                            TextOrCData::from_xml_with_reader(bufs, reader)?
                                    }

                                    "ttl" => ttl = NumberData::from_xml_with_reader(bufs, reader)?,

                                    "image" => {
                                        image =
                                            Some(ChannelImage::from_xml_with_reader(bufs, reader)?)
                                    }

                                    "item" => {
                                        let item = ChannelItem::from_xml_with_reader(bufs, reader)?;
                                        posts.push(item);
                                    }

                                    "copyright" => {
                                        copyright = TextOrCData::from_xml_with_reader(bufs, reader)?
                                    }

                                    _ => {
                                        SkipThisElement::from_xml_with_reader(bufs, reader)?;
                                    }
                                },

                                Ok(Event::Eof | Event::End(_)) => break,
                                Ok(_) => (),
                                Err(e) => return Err(e),
                            }
                            cbuf.clear()
                        }

                        console_log!("item count: {}", posts.len());
                    }

                    _ => (),
                },

                Ok(Event::Eof | Event::End(_)) => break,
                Ok(_) => (),

                Err(e) => panic!("Error at position {}: {:?}", reader.buffer_position(), e),
            }
            buf.clear();
        }

        Ok(Self {
            version,
            title,
            description,
            url,
            atom_link,
            language,
            web_master,
            generator,
            last_build_date,
            ttl,
            image,
            posts,
            copyright,
        })
    }
}

impl FromStr for RSSChannel {
    type Err = fast_xml::Error;

    /// It takes a string, parses it as XML, and returns a `fast_xml::Result<RSSChannel>`
    ///
    /// Arguments:
    ///
    /// * `text`: The XML text to parse.
    ///
    /// Returns:
    ///
    /// A Result<RSSChannel, fast_xml::Error>
    fn from_str(text: &str) -> fast_xml::Result<RSSChannel> {
        let bufs = BufPool::new(16, 2048);

        Self::from_xml_with_str(&bufs, text)
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
        serde_wasm_bindgen::to_value(&self).unwrap()
    }

    #[wasm_bindgen(js_name = postsLen)]
    pub fn posts_len(&self) -> usize {
        self.posts.len()
    }

    #[wasm_bindgen(js_name = posts)]
    pub fn posts(&self) -> JsValue {
        serde_wasm_bindgen::to_value(&self.posts).unwrap()
    }
}

#[wasm_bindgen]
#[derive(Debug, Serialize, Deserialize)]
#[serde(rename = "image")]
pub struct ChannelImage {
    url: Option<String>,
}

impl FromXmlWithReader for ChannelImage {
    fn from_xml_with_reader<B: std::io::BufRead>(
        bufs: &BufPool,
        reader: &mut Reader<B>,
    ) -> fast_xml::Result<Self> {
        let mut url = None;

        let mut buf = bufs.pop();

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Start(ref e)) => match reader.decode(e.local_name())? {
                    "url" => url = TextOrCData::from_xml_with_reader(bufs, reader)?,
                    _ => {
                        SkipThisElement::from_xml_with_reader(bufs, reader)?;
                    }
                },
                Ok(Event::End(_) | Event::Eof) => break,
                Ok(_) => (),

                Err(e) => return Err(e),
            }
            buf.clear();
        }

        Ok(Self { url })
    }
}
