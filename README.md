# Ressu

[![build](https://github.com/sdttttt/ressu/actions/workflows/build.yml/badge.svg?branch=master)](https://github.com/sdttttt/ressu/actions/workflows/build.yml)

This is a free, modern **RSS Reader**, built using React in WebView(Tauri).

*Of course, it's free.*


> Ressu primary target is the Windows platform and then the Linux platform.


## Build

You need the following environment:

- Rust
- Node.js
- pnpm (node package manager)
- wasm-pack

After WASM compile complete, need **reinstall node_modules** to found wasm library. It is imported as link in package.json.

## TODO

**Stage 1:**

- [x] RSS Parse.
- [x] Local Storage.
- [ ] UI Design.

**Stage 2:**

- [ ] Complete settings options. 

## LICENSE

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or distribute this software, either in source code form or as a compiled binary, for any purpose, commercial or non-commercial, and by any means.

The author disclaims copyright to this source code. use `UNLICENSE`.
