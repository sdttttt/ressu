#[macro_use]
mod utils;
mod constants;
mod buf;

pub mod feed;
pub mod item;
pub mod js_bind;

use std::io::BufRead;

use crate::buf::BufPool;

use wasm_bindgen::prelude::*;

// When the `wee_alloc` feature is enabled, use `wee_alloc` as the global
// allocator.
#[cfg(feature = "wee_alloc")]
#[global_allocator]
static ALLOC: wee_alloc::WeeAlloc = wee_alloc::WeeAlloc::INIT;


#[wasm_bindgen(start, js_name = onStart)]
pub fn on_start() {
    console_log!("WASM starting ...");
    let window = web_sys::window().expect("should have a window in this context");
    window.document().expect("window should have a document");
}

#[wasm_bindgen]
pub fn hello() {
   console_log!("from WASM greet!");
}

/// A trait that is implemented by the structs that are used to parse the XML.
pub trait FromXmlWithStr: Sized {
    /// A function that takes a BufPool and a string and returns a fast_xml::Result<Self>
	fn from_xml_with_str(
        bufs: &BufPool,
        text: &str
    ) -> fast_xml::Result<Self>;
}

pub trait FromXmlWithReader: Sized {
    fn from_xml_with_reader<B: BufRead>(
        bufs: &BufPool,
        reader: &mut XmlReader<B>
    ) -> fast_xml::Result<Self>;
}

struct SkipThisElement;

use fast_xml::events::Event as XmlEvent;
use fast_xml::Reader as XmlReader;

impl FromXmlWithReader for SkipThisElement {

    /// "Skip the current element and all its children."
	/// 
	/// The function is called `from_xml_with_reader` because it's part of a trait called
	/// `FromXmlWithReader` that's implemented for all types that can be deserialized from XML
	/// 
	/// Arguments:
	/// 
	/// * `bufs`: &BufPool
	/// * `reader`: &mut XmlReader<B>
	/// 
	/// Returns:
	/// 
	/// A SkipThisElement
    #[inline]
	fn from_xml_with_reader<B: BufRead>(
        bufs: &BufPool,
        reader: &mut XmlReader<B>,
    ) -> fast_xml::Result<Self> {
        let mut buf = bufs.pop();
        let mut depth = 1usize;
        loop {
            match reader.read_event(&mut buf) {
                Ok(XmlEvent::Start(_)) => depth += 1,
                Ok(XmlEvent::End(_)) if depth == 1 => break,
                Ok(XmlEvent::End(_)) => depth -= 1,
                Ok(XmlEvent::Eof) => break, // just ignore EOF
                Err(err) => return Err(err),
                _ => (),
            }
            buf.clear();
        }
        Ok(SkipThisElement)
    }
}
