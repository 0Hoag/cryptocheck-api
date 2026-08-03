package sites

import (
	"context"
	"errors"
	"testing"
)

func TestExtractRSSImageURL(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "embedded HTTPS image",
			description: `<p>Summary</p><img src="https://images.example.com/preview.png?width=1200">`,
			want:        "https://images.example.com/preview.png?width=1200",
		},
		{
			name:        "non HTTPS image is rejected",
			description: `<img src="http://images.example.com/preview.png">`,
		},
		{
			name:        "missing image",
			description: `<p>Summary only</p>`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := extractRSSImageURL(test.description); got != test.want {
				t.Fatalf("extractRSSImageURL() = %q, want %q", got, test.want)
			}
		})
	}
}

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
