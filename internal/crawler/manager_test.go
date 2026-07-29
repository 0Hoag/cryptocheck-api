package crawler

import (
	"testing"
	"time"
)

func TestNormalizeArticlesSortsNewestAndDeduplicatesURLs(t *testing.T) {
	now := time.Date(2026, 7, 29, 0, 0, 0, 0, time.UTC)
	articles := normalizeArticles([]Article{
		{Title: "Old", SourceURL: "https://example.com/old", PublishedAt: now.Add(-2 * time.Hour)},
		{Title: "Newest", SourceURL: "https://example.com/new", PublishedAt: now},
		{Title: "Duplicate", SourceURL: "https://example.com/new", PublishedAt: now.Add(-time.Hour)},
		{Title: "Middle", SourceURL: "https://example.com/middle", PublishedAt: now.Add(-time.Hour)},
	})

	if len(articles) != 3 {
		t.Fatalf("normalized article count = %d, want 3", len(articles))
	}
	if articles[0].Title != "Newest" || articles[1].Title != "Middle" || articles[2].Title != "Old" {
		t.Fatalf("articles were not sorted newest-first: %+v", articles)
	}
}
