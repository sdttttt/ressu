package sanitize_test

import (
	"strings"
	"testing"

	"github.com/sdttttt/ressu/internal/sanitize"
)

func TestHTMLRemovesForbiddenElements(t *testing.T) {
	input := `<p>正文</p>` +
		`<script>alert(1)</script>` +
		`<style>p{color:red}</style>` +
		`<iframe src="https://evil.example"></iframe>` +
		`<object data="payload"></object><object>fallback</object>` +
		`<embed src="payload">` +
		`<form action="/subscribe"><p>form text</p></form>` +
		`<input type="text">` +
		`<button><p>button text</p></button>`
	output := sanitize.HTML(input, "https://example.com/post")

	for _, unsafe := range []string{
		"<script", "alert(1)", "<style", "color:red", "<iframe", "<object", "<embed",
		"<form", "<input", "<button",
	} {
		if strings.Contains(output, unsafe) {
			t.Fatalf("%q survived: %s", unsafe, output)
		}
	}
	for _, text := range []string{"正文", "form text", "button text"} {
		if !strings.Contains(output, text) {
			t.Fatalf("text %q was lost: %s", text, output)
		}
	}
}

func TestHTMLAllowsSafeHTMLAndDropsUnsafeAttributes(t *testing.T) {
	input := `<article class="post" tabindex="0" data-feed="rss" aria-label="Article" ` +
		`onclick="alert(1)" style="color:red"><table><tr><td>cell</td></tr></table>` +
		`<custom-tag>custom text</custom-tag><template><p>template body</p></template>` +
		`<video controls><source src="/media.mp4">fallback</video></article>`
	output := sanitize.HTML(input, "https://example.com/post")

	for _, safe := range []string{
		`<article aria-label="Article" class="post" data-feed="rss" tabindex="0">`,
		"<td>cell</td>", "custom text",
		`<video controls="">`, `src="https://example.com/media.mp4"`,
	} {
		if !strings.Contains(output, safe) {
			t.Fatalf("safe output %q was lost: %s", safe, output)
		}
	}
	for _, unsafe := range []string{
		"onclick", "alert(1)", "color:red", "<custom-tag", "template body",
		"<template",
	} {
		if strings.Contains(output, unsafe) {
			t.Fatalf("%q survived: %s", unsafe, output)
		}
	}
}

func TestHTMLCompletesURLsAndAddsReadingBehavior(t *testing.T) {
	input := `<a href="/next">next</a><a href="mailto:hi@example.com">mail</a>` +
		`<img src="cover.png" width="800" height="600">` +
		`<source src="/audio.ogg">`
	output := sanitize.HTML(input, "https://example.com/posts/1")

	for _, safe := range []string{
		`<a href="https://example.com/next" rel="noopener noreferrer" target="_blank">next</a>`,
		`<a href="mailto:hi@example.com" rel="noopener noreferrer" target="_blank">mail</a>`,
		`<img decoding="async" loading="lazy" src="https://example.com/posts/cover.png"/>`,
		`<source src="https://example.com/audio.ogg"/>`,
	} {
		if !strings.Contains(output, safe) {
			t.Fatalf("safe output %q was lost: %s", safe, output)
		}
	}
	if strings.Contains(output, "width=") || strings.Contains(output, "height=") {
		t.Fatalf("image dimensions survived: %s", output)
	}
}

func TestHTMLRemovesUnsafeURIsAndKeepsImageDataURIs(t *testing.T) {
	input := `<a href="javascript:alert(1)">bad link</a>` +
		`<a href="data:text/html,bad">data link</a>` +
		`<img src="javascript:alert(1)">` +
		`<img src="data:image/png;base64,aGVsbG8=">` +
		`<source src="data:text/html,bad">`
	output := sanitize.HTML(input, "https://example.com/post")

	for _, unsafe := range []string{
		`href="javascript:`, `href="data:text/html`,
	} {
		if strings.Contains(output, unsafe) {
			t.Fatalf("%q survived: %s", unsafe, output)
		}
	}
	if !strings.Contains(output, `src="data:image/png;base64,aGVsbG8="`) {
		t.Fatalf("safe image data URI was lost: %s", output)
	}
	if strings.Contains(output, `<source src="data:text/html`) {
		t.Fatalf("unsafe source data URI survived: %s", output)
	}
}

func TestHTMLRemovesComments(t *testing.T) {
	input := `<!-- safe comment --><!-- <script>alert(1)</script> --><p>text</p>`
	output := sanitize.HTML(input, "https://example.com/post")

	for _, unsafe := range []string{"safe comment", "alert(1)", "<script", "<!--"} {
		if strings.Contains(output, unsafe) {
			t.Fatalf("comment survived: %s", output)
		}
	}
}

func TestHTMLOmitsDocumentWrapperAndIsIdempotent(t *testing.T) {
	input := `<!doctype html><html><head><title>hidden</title></head>` +
		`<body><p>visible</p></body></html>`
	first := sanitize.HTML(input, "https://example.com/")
	second := sanitize.HTML(first, "https://example.com/")

	if strings.Contains(first, "<html>") || strings.Contains(first, "<head>") ||
		strings.Contains(first, "hidden") {
		t.Fatalf("document wrapper survived: %s", first)
	}
	if first != second {
		t.Fatalf("sanitization was not idempotent:\nfirst=%s\nsecond=%s", first, second)
	}
}
