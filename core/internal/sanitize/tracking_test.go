package sanitize

import (
	"strings"
	"testing"
)

func TestHTMLStripsTrackingParamsFromLinks(t *testing.T) {
	input := `<p><a href="https://target.example/page?utm_source=newsletter&amp;utm_medium=email&amp;id=42&amp;fbclid=abc&amp;spm=a2o1">go</a></p>`
	output := HTML(input, "https://source.example/post", true)

	link := `href="https://target.example/page?id=42"`
	if !strings.Contains(output, link) {
		t.Errorf("stripped link = %s, want %s", output, link)
	}
	for _, tracker := range []string{"utm_source", "utm_medium", "fbclid", "spm"} {
		if strings.Contains(output, tracker) {
			t.Errorf("tracking parameter %q survived: %s", tracker, output)
		}
	}
}

func TestHTMLStripKeepsSelfReferenceParams(t *testing.T) {
	input := `<a href="https://target.example/p?ref=source.example">self</a>` +
		`<a href="https://target.example/p?ref=other.example">other</a>`
	output := HTML(input, "https://source.example/post", true)

	if strings.Contains(output, "ref=source.example") {
		t.Errorf("self-referencing ref survived: %s", output)
	}
	if !strings.Contains(output, "ref=other.example") {
		t.Errorf("foreign ref was stripped: %s", output)
	}
}

func TestHTMLStripIsCaseInsensitiveAndDecodesNames(t *testing.T) {
	input := `<a href="https://target.example/p?UTM_SOURCE=x&amp;%66bclid=y&amp;id=1">go</a>`
	output := HTML(input, "https://source.example/post", true)

	if strings.Contains(strings.ToLower(output), "utm_source") || strings.Contains(output, "fbclid") {
		t.Errorf("encoded or uppercase trackers survived: %s", output)
	}
	if !strings.Contains(output, "id=1") {
		t.Errorf("ordinary parameter was stripped: %s", output)
	}
}

func TestHTMLStripDisabledLeavesLinksAlone(t *testing.T) {
	input := `<a href="https://target.example/p?utm_source=x">go</a>`
	output := HTML(input, "https://source.example/post", false)

	if !strings.Contains(output, "utm_source=x") {
		t.Errorf("strip=false changed the link: %s", output)
	}
}

func TestHTMLStripSkipsLinksWithoutHostOrQuery(t *testing.T) {
	input := `<a href="#anchor">anchor</a><a href="mailto:a@b.example">mail</a>`
	output := HTML(input, "https://source.example/post", true)

	// absolute() resolves the anchor against the base, as it always has;
	// stripping must not remove anything from either address.
	if !strings.Contains(output, "https://source.example/post#anchor") ||
		!strings.Contains(output, "mailto:a@b.example") {
		t.Errorf("non-HTTP links were touched: %s", output)
	}
}

func TestHTMLStripIsIdempotent(t *testing.T) {
	input := `<a href="https://target.example/p?utm_source=x&amp;id=1">go</a>`
	first := HTML(input, "https://source.example/post", true)
	second := HTML(first, "https://source.example/post", true)

	if first != second {
		t.Fatalf("stripping was not idempotent:\nfirst=%s\nsecond=%s", first, second)
	}
}

func TestStripTrackingParamsWithoutSiteContext(t *testing.T) {
	cleaned := StripTrackingParams("https://target.example/p?utm_source=x&ref=anything", "")
	if strings.Contains(cleaned, "utm_source") {
		t.Errorf("utm_source survived without site context: %s", cleaned)
	}
	if !strings.Contains(cleaned, "ref=anything") {
		t.Errorf("ref lost its value-based exception without site context: %s", cleaned)
	}
}
