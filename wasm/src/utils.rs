use fast_xml::{
    events::{attributes::Attributes, Event},
    Reader,
};

use std::{borrow::Cow};

use crate::buf::BufPool;
use crate::FromXmlWithReader;

#[allow(dead_code)]
pub fn set_panic_hook() {
    // When the `console_error_panic_hook` feature is enabled, we can call the
    // `set_panic_hook` function at least once during initialization, and then
    // we will get better error messages if our code ever panics.

    // For more details see
    // https://github.com/rustwasm/console_error_panic_hook#readme
    #[cfg(feature = "console_error_panic_hook")]
    console_error_panic_hook::set_once();
}

#[macro_export]
macro_rules! console_log  {
	($($t:tt)*) => (crate::js_bind::log(&format_args!($($t)*).to_string()))
}

pub type TextOrCData = Option<String>;
pub type NumberData = Option<usize>;

impl FromXmlWithReader for TextOrCData {
    /// It reads the next event from the reader, and if it's a text or CDATA event, it returns the text.
    /// Otherwise, it returns an empty string
    ///
    /// Arguments:
    ///
    /// * `reader`: &mut Reader<B>
    /// * `bufs`: A BufPool is a pool of buffers that are used to read the XML.
    ///
    /// Returns:
    ///
    /// A string
    fn from_xml_with_reader<B: std::io::BufRead>(
        bufs: &BufPool,
        reader: &mut Reader<B>,
    ) -> fast_xml::Result<Self> {
        let mut buf = bufs.pop();

        let mut text = None;

        loop {
            match reader.read_event(&mut buf) {
                Ok(Event::Text(ref e) | Event::CData(ref e)) => {
                    text = Some(reader.decode(e)?.to_string())
                }
                Ok(Event::End(_) | Event::Eof) => break,
                Ok(_) => {}
                Err(e) => return Err(e),
            };
            buf.clear();
        }

        Ok(text)
    }
}

impl FromXmlWithReader for NumberData {
    /// It reads the XML file, and if it finds a number, it returns it
		/// 
		/// Arguments:
		/// 
		/// * `bufs`: &BufPool,
		/// * `reader`: &mut Reader<B>
		/// 
		/// Returns:
		/// 
		/// A Result<Option<usize>>
		fn from_xml_with_reader<B: std::io::BufRead>(
        bufs: &BufPool,
        reader: &mut Reader<B>,
    ) -> fast_xml::Result<Self> {
			let mut buf = bufs.pop();
			let mut number = None;

			loop {
				match reader.read_event(&mut buf) {
						Ok(Event::Text(ref e) | Event::CData(ref e)) => {
							let number_str = reader.decode(e)?;
							number = Some(
									match number_str.parse::<usize>() {
										Err(_) => return Err(fast_xml::Error::UnexpectedToken(
											(format!("{} to number failed!", number_str))
											.to_string())),
										Ok(n) => n,
									}
								)
						}
						Ok(Event::End(_) | Event::Eof) => break,
						Ok(_) => {}
						Err(e) => return Err(e),
				};
				buf.clear();
		}

		Ok(number)
		}
}

///
/// from attributes get key str.
/// It takes a `Reader` and `Attributes` and a `key` and returns a `Result` of an `Option` of a `String`
///
/// Arguments:
///
/// * `reader`: &Reader<B> - The reader that is reading the XML file.
/// * `attrs`: Attributes<'a>
/// * `key`: The key of the attribute you want to get the value of.
///
/// Returns:
///
/// A function that takes a reader, attributes, and a key and returns a Result of an Option of a String.
pub fn attrs_get_str<'a, B: std::io::BufRead>(
    reader: &Reader<B>,
    attrs: Attributes<'a>,
    key: &'a str,
) -> fast_xml::Result<Option<String>> {
    let mut value = None;

    for attribute in attrs {
        let attribute = attribute?;

        if attribute.key != key.as_bytes() {
            continue;
        }

        value = Some(
            reader
                .decode(if let Cow::Borrowed(s) = attribute.value {
                    s
                } else {
                    unreachable!();
                })?
                .to_string(),
        );
    }
    Ok(value)
}

/// It reads the XML file from the current position to the end of the current tag, and returns the text
/// between the current position and the end of the current tag
///
/// Arguments:
///
/// * `reader`: &mut Reader<B>
/// * `bufs`: a pool of buffers that are used to read the XML file.
/// * `tag`: the tag name of the node you want to get
/// * `origin_text`: the original text of the XML file
///
/// Returns:
///
/// A string.
#[allow(dead_code)]
pub fn reader_get_sub_node_str<B: std::io::BufRead>(
    reader: &mut Reader<B>,
    bufs: &BufPool,
    tag: &str,
    origin_text: &str,
) -> fast_xml::Result<String> {
    let mut buf = bufs.pop();

    let start_position = reader.buffer_position();

    reader.check_end_names(false);
    loop {
        match reader.read_event(&mut buf) {
            Ok(Event::End(ref e)) => {
                if reader.decode(e.name())? == "item" {
                    break;
                }
            }
            Ok(Event::Eof) => break,
            _ => (),
        }
        buf.clear();
    }
    reader.check_end_names(true);

    let end_position = reader.buffer_position();

    let end_tag_len = format!("</{}>", tag).len();

    let text_string = origin_text.to_string();
    let item_slice = &text_string.as_bytes()[start_position..end_position - end_tag_len];

    console_log!(
        "tag: {}, start: {}, end: {}, size: {}",
        tag,
        start_position,
        end_position - end_tag_len,
        end_position - end_tag_len - start_position
    );

    Ok(unsafe { String::from_utf8_unchecked(item_slice.into()) })
}
