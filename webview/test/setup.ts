// jsdom lacks the browser APIs that the UI primitives reach for.

if (!globalThis.matchMedia) {
  globalThis.matchMedia = ((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: () => {},
    removeEventListener: () => {},
    addListener: () => {},
    removeListener: () => {},
    dispatchEvent: () => false,
  })) as unknown as typeof globalThis.matchMedia;
}

if (!globalThis.ResizeObserver) {
  globalThis.ResizeObserver = class {
    observe() {}
    unobserve() {}
    disconnect() {}
  } as unknown as typeof globalThis.ResizeObserver;
}

if (!Element.prototype.scrollIntoView) {
  Element.prototype.scrollIntoView = () => {};
}

if (!Element.prototype.scrollBy) {
  Element.prototype.scrollBy = () => {};
}

// The webview never talks to the network directly (data comes through the
// MyGO bridge), and tests must not reach out either: make any accidental
// attempt fail loudly instead of hanging or silently resolving.
for (const name of ["fetch", "XMLHttpRequest", "WebSocket", "EventSource"] as const) {
  const owner = globalThis as unknown as Record<string, unknown>;
  owner[name] = () => {
    throw new Error(`network is banned in the webview: ${name} was called`);
  };
}
