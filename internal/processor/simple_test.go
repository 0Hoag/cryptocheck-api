package processor

import (
	"errors"
	"testing"
)

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

func TestTranslatedOrFallback(t *testing.T) {
	if got, ok := translatedOrFallback("Bản dịch", "Original", nil); !ok || got != "Bản dịch" {
		t.Fatalf("successful translation = (%q, %t), want translated result", got, ok)
	}
	if got, ok := translatedOrFallback("", "Original", nil); ok || got != "Original" {
		t.Fatalf("empty translation = (%q, %t), want fallback", got, ok)
	}
	if got, ok := translatedOrFallback("ignored", "Original", errors.New("provider unavailable")); ok || got != "Original" {
		t.Fatalf("failed translation = (%q, %t), want fallback", got, ok)
	}
}
