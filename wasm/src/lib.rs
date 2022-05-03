#[macro_use]
mod utils;
mod js_bind;
mod constants;

pub mod feed;
pub mod item;


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
