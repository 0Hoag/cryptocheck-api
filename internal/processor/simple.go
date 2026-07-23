package processor

import (
	"context"
	"html"
	"regexp"
	"strings"

	"github.com/0Hoag/cryptocheck-api/internal/crawler"
	"github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/bregydoc/gtranslate"
)

type SimpleProcessor struct {
	l log.Logger
}

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)
var spaceBeforePunctuationPattern = regexp.MustCompile(`\s+([.,!?;:])`)

// cleanArticleText turns RSS HTML and crawler output into safe plain text for
// translation and Markdown rendering. Cover images are stored separately.
func cleanArticleText(value string) string {
	value = htmlTagPattern.ReplaceAllString(value, " ")
	value = html.UnescapeString(value)
	value = strings.Join(strings.Fields(value), " ")
	return spaceBeforePunctuationPattern.ReplaceAllString(value, "$1")
}

func articleExcerpt(content, fallback string) string {
	text := cleanArticleText(content)
	if len(text) < 80 {
		text = cleanArticleText(fallback)
	}
	if len(text) <= 900 {
		return text
	}
	cut := text[:900]
	if lastSentence := strings.LastIndexAny(cut, ".!?"); lastSentence > 400 {
		return cut[:lastSentence+1]
	}
	return strings.TrimSpace(cut) + "…"
}

func NewSimpleProcessor(l log.Logger) *SimpleProcessor {
	return &SimpleProcessor{l: l}
}

func (p *SimpleProcessor) Process(ctx context.Context, article crawler.Article) (ProcessedContent, error) {
	// 1. Translate Title
	titleVi, err := gtranslate.TranslateWithParams(
		article.Title,
		gtranslate.TranslationParams{
			From: "en",
			To:   "vi",
		},
	)
	if err != nil {
		p.l.Errorf(ctx, "Failed to translate title: %v", err)
		p.l.Errorf(ctx, "processor.simple.Process.TranslateTitle: %v", err)
		return ProcessedContent{}, err
	}

	// Prefer an excerpt from the crawled article body. RSS descriptions often
	// contain tracking HTML and a duplicated cover image.
	summary := articleExcerpt(article.Content, article.Summary)

	// If still no summary, use title
	if summary == "" {
		summary = article.Title
	}

	// Translate summary
	summaryVi, err := gtranslate.TranslateWithParams(
		summary,
		gtranslate.TranslationParams{
			From: "en",
			To:   "vi",
		},
	)
	if err != nil {
		p.l.Errorf(ctx, "processor.simple.Process.TranslateSummary: %v", err)
		return ProcessedContent{}, err
	}

	return ProcessedContent{
		OriginalTitle:         article.Title,
		TranslatedTitle:       titleVi,
		OriginalSummary:       summary,
		TranslatedSummary:     summaryVi,
		TranslatedFullContent: summaryVi,
		Content:               summary,
		SourceURL:             article.SourceURL,
		ImageURL:              article.ImageURL,
	}, nil
}
