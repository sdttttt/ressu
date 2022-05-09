# Wasm for Ressu

Part of the functionality of Ressu's WASM implementation, responsible for the RSS/XML parse and process work. 

Part of the code comes from [github.com/iovxw/rssbot](https://github.com/iovxw/rssbot).


# Benchmark

from wasm_bindgen_test:

- It takes 27s to process about 3000 items, each of which is about 2.5KB in size. (data/JUMP_NS_ALL_RSS2.0.xml)
- It takes 0.36s to process about 100 items, each of which is about 2.5kb in size. (data/JUMP_PS5_ALL_RSS2.0.xml)
- It takes 3.3s to process about 1600 items, each of which is about 0.5kb in size. (data/QIDIAN_RSS2.0.xml)
- It takes 0.1s to process about 14 items, each of which is about 10kb in size. (data/ELS_RSS2.0.xml)
- It takes 0.007s to process about 20 items, each of which is about 4kb in size. (data/3DM_RSS2.0.xml)
