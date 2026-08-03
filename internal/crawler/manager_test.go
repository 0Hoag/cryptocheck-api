package crawler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

type noopLogger struct{}

var _ log.Logger = noopLogger{}

func (noopLogger) Debug(context.Context, ...any)          {}
func (noopLogger) Debugf(context.Context, string, ...any) {}
func (noopLogger) Info(context.Context, ...any)           {}
func (noopLogger) Infof(context.Context, string, ...any)  {}
func (noopLogger) Warn(context.Context, ...any)           {}
func (noopLogger) Warnf(context.Context, string, ...any)  {}
func (noopLogger) Error(context.Context, ...any)          {}
func (noopLogger) Errorf(context.Context, string, ...any) {}
func (noopLogger) Fatal(context.Context, ...any)          {}
func (noopLogger) Fatalf(context.Context, string, ...any) {}

type testCrawler struct {
	name     string
	articles []Article
	err      error
}

func (c testCrawler) Name() string { return c.name }

func (c testCrawler) Crawl(context.Context) ([]Article, error) {
	return c.articles, c.err
}

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

func TestManagerReturnsErrorOnlyWhenEveryCrawlerFails(t *testing.T) {
	manager := NewManager(noopLogger{})
	manager.Register(testCrawler{name: "failed-one", err: errors.New("unavailable")})
	manager.Register(testCrawler{name: "failed-two", err: errors.New("blocked")})

	articles, err := manager.Run(context.Background())
	if err == nil {
		t.Fatal("Run() error = nil, want an error when all crawlers fail")
	}
	if articles != nil {
		t.Fatalf("Run() articles = %#v, want nil", articles)
	}
}

func TestManagerKeepsHealthySourceWhenAnotherCrawlerFails(t *testing.T) {
	manager := NewManager(noopLogger{})
	manager.Register(testCrawler{name: "failed", err: errors.New("unavailable")})
	manager.Register(testCrawler{name: "healthy", articles: []Article{{Title: "Fresh", SourceURL: "https://example.com/fresh"}}})

	articles, err := manager.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error = %v, want nil while one crawler is healthy", err)
	}
	if len(articles) != 1 || articles[0].Title != "Fresh" {
		t.Fatalf("Run() articles = %#v, want healthy crawler output", articles)
	}
}
