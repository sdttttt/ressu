# Wasm for Ressu

Part of the functionality of Ressu's WASM implementation, responsible for the RSS/XML parse and process work. 

Part of the code comes from [github.com/iovxw/rssbot](https://github.com/iovxw/rssbot).


# Benchmark

Under the blessing of WASM of Rust, the memory copy is optimized and the terrifying speed of RSS parsing is achieved.

from wasm_bindgen_test:

- It takes **446ms** to process about **3000 items**, each of which is about **2.5KB** in size. (data/JUMP_NS_ALL_RSS2.0.xml)
- It takes **16ms** to process about **100 items**, each of which is about **2.5kb** in size. (data/JUMP_PS5_ALL_RSS2.0.xml)
- It takes **67ms** to process about **1600 items**, each of which is about **0.5kb** in size. (data/QIDIAN_RSS2.0.xml)


