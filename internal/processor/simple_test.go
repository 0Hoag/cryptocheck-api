package processor

import "testing"

func TestArticleExcerptStripsRSSHTMLAndPrefersBody(t *testing.T) {
	body := "First body paragraph with useful context. Second body paragraph provides more detail for readers."
	rss := `<p><img src="https://example.com/image.png">RSS teaser should not be selected.</p>`
	got := articleExcerpt(body, rss)
	if got != body {
		t.Fatalf("excerpt = %q, want body text", got)
	}
}

func TestCleanArticleTextRemovesImageMarkup(t *testing.T) {
	got := cleanArticleText(`<p><img src="x.png">Crypto &amp; markets <strong>update</strong>.</p>`)
	if got != "Crypto & markets update." {
		t.Fatalf("cleanArticleText = %q", got)
	}
}
