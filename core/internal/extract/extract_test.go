package extract_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sdttttt/ressu/core/internal/extract"
)

const articlePage = `<!DOCTYPE html>
<html>
<head><title>Example article</title></head>
<body>
<nav><a href="/menu">Menu</a></nav>
<article>
<h1>Example article</h1>
<p>The readable body of the article, long enough for the extractor to
pick this node over the surrounding chrome.</p>
<p>It keeps going with a second paragraph so the candidate node scores
well above the navigation boilerplate above and beyond any doubt.</p>
<script>alert("x")</script>
<img src="/image.png" alt="diagram">
</article>
<footer>Footer boilerplate</footer>
</body>
</html>`

func TestArticleExtractsReadableBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/article" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(articlePage))
	}))
	defer server.Close()

	content, err := extract.Article(context.Background(), server.URL+"/article", false)
	if err != nil {
		t.Fatalf("Article: %v", err)
	}
	if !strings.Contains(content, "readable body of the article") {
		t.Errorf("extracted content = %q; want the article body", content)
	}
	if strings.Contains(content, "alert(") {
		t.Errorf("extracted content keeps a script: %q", content)
	}
	if !strings.Contains(content, server.URL+"/image.png") {
		t.Errorf("extracted content = %q; want the image resolved against the page", content)
	}
}

func TestArticleRejectsBadAddresses(t *testing.T) {
	for _, raw := range []string{"", "ftp://example.com/a", "javascript:alert(1)", "/relative"} {
		if _, err := extract.Article(context.Background(), raw, false); err == nil {
			t.Errorf("Article(%q) succeeded; want an error", raw)
		}
	}
}

func TestArticleReportsServerErrors(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()

	if _, err := extract.Article(context.Background(), server.URL+"/missing", false); err == nil {
		t.Fatal("Article succeeded on a 404; want an error")
	}
}

func TestArticleReportsUnextractablePages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("<html><head><title>x</title></head><body></body></html>"))
	}))
	defer server.Close()

	_, err := extract.Article(context.Background(), server.URL, false)
	if !errors.Is(err, extract.ErrExtraction) {
		t.Fatalf("error = %v, want extract.ErrExtraction", err)
	}
}
