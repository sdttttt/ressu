// Package textmetrics counts the readable text in article HTML.
package textmetrics

import (
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// ignoredElements contain content that a reader does not display as body
// text.
var ignoredElements = map[string]bool{
	"iframe":   true,
	"math":     true,
	"noscript": true,
	"object":   true,
	"script":   true,
	"style":    true,
	"svg":      true,
	"template": true,
}

// blockElements separate words instead of joining text across markup.
var blockElements = map[string]bool{
	"article":    true,
	"aside":      true,
	"blockquote": true,
	"br":         true,
	"caption":    true,
	"dd":         true,
	"div":        true,
	"dl":         true,
	"dt":         true,
	"fieldset":   true,
	"figcaption": true,
	"figure":     true,
	"footer":     true,
	"h1":         true,
	"h2":         true,
	"h3":         true,
	"h4":         true,
	"h5":         true,
	"h6":         true,
	"header":     true,
	"hr":         true,
	"li":         true,
	"main":       true,
	"nav":        true,
	"ol":         true,
	"p":          true,
	"pre":        true,
	"section":    true,
	"table":      true,
	"tbody":      true,
	"td":         true,
	"tfoot":      true,
	"th":         true,
	"thead":      true,
	"tr":         true,
	"ul":         true,
}

// Count reports the reading units in HTML or plain text. CJK characters count
// individually; contiguous letters and digits count as one unit.
func Count(markup string) int {
	if markup == "" {
		return 0
	}

	tokenizer := html.NewTokenizer(strings.NewReader(markup))
	var ignored []string
	count := 0
	inWord := false
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return count
		case html.TextToken:
			if len(ignored) > 0 {
				continue
			}
			countRunes(string(tokenizer.Text()), &count, &inWord)
		case html.StartTagToken:
			name := tagName(tokenizer)
			if ignoredElements[name] {
				ignored = append(ignored, name)
				continue
			}
			if blockElements[name] {
				inWord = false
			}
		case html.EndTagToken:
			name := tagName(tokenizer)
			if len(ignored) > 0 && ignored[len(ignored)-1] == name {
				ignored = ignored[:len(ignored)-1]
				continue
			}
			if blockElements[name] {
				inWord = false
			}
		}
	}
}

// countRunes adds the reading units in one run of visible text.
func countRunes(text string, count *int, inWord *bool) {
	for _, value := range text {
		switch {
		case isCJK(value):
			*count++
			*inWord = false
		case unicode.IsLetter(value) || unicode.IsDigit(value):
			if !*inWord {
				*count++
				*inWord = true
			}
		default:
			*inWord = false
		}
	}
}

// tagName returns the lowercased name of the current tag token.
func tagName(tokenizer *html.Tokenizer) string {
	name, _ := tokenizer.TagName()
	return strings.ToLower(string(name))
}

// isCJK reports whether a rune is Chinese, Japanese kana, or Korean.
func isCJK(value rune) bool {
	return unicode.Is(unicode.Han, value) ||
		unicode.Is(unicode.Hiragana, value) ||
		unicode.Is(unicode.Katakana, value) ||
		unicode.Is(unicode.Hangul, value)
}
