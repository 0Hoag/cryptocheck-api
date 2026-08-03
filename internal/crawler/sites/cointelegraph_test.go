package sites

import (
	"context"
	"errors"
	"testing"
)

func TestCoinTelegraphCrawlerHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	articles, err := NewCoinTelegraphCrawler().Crawl(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Crawl() error = %v, want context.Canceled", err)
	}
	if articles != nil {
		t.Fatalf("Crawl() articles = %#v, want nil", articles)
	}
}

func TestNewCoinTelegraphCrawlerCreatesIndependentInstances(t *testing.T) {
	first := NewCoinTelegraphCrawler()
	second := NewCoinTelegraphCrawler()
	if first == second {
		t.Fatal("NewCoinTelegraphCrawler() returned the same crawler instance")
	}
}
