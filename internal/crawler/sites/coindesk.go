package sites

import (
	"context"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/crawler"
	"github.com/gocolly/colly/v2"
)

type CoindeskCrawler struct {
}

// The worker publishes at most ten newest articles per run. Keeping a modest
// source-level ceiling avoids opening every link on the CoinDesk homepage
// before that publish cap is applied.
const maxCoindeskArticles = 20

func parsePublishedAt(value string, fallback time.Time) time.Time {
	publishedAt, err := time.Parse(time.RFC3339, strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return publishedAt
}

func NewCoindeskCrawler() *CoindeskCrawler {
	return &CoindeskCrawler{}
}

func (c *CoindeskCrawler) Name() string {
	return "coindesk"
}

func (c *CoindeskCrawler) Crawl(ctx context.Context) ([]crawler.Article, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var articles []crawler.Article

	// Create collector by libary Colly
	collector := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"),
	)

	// Coindesk structure often changes. We target common article wrappers.
	// This selector targets the main news texts.
	// Note: We might need to refine this based on the actual HTML structure.
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		title := strings.TrimSpace(e.Text)

		// Filter relevant links
		if title == "" || len(strings.Split(title, " ")) < 3 { // Skip short link texts
			return
		}

		// Ensure full URL
		if !strings.HasPrefix(link, "http") {
			link = "https://www.coindesk.com" + link
		}

		// Basic filter to ensure it looks like an article
		if strings.Contains(link, "/business/") ||
			strings.Contains(link, "/markets/") ||
			strings.Contains(link, "/policy/") ||
			strings.Contains(link, "/tech/") ||
			strings.Contains(link, "/opinion/") {

			// Deduplicate if already added (simple check)
			for _, a := range articles {
				if a.SourceURL == link {
					return
				}
			}
			if len(articles) >= maxCoindeskArticles || ctx.Err() != nil {
				return
			}

			// Create detail collector to fetch image and full content
			detailCollector := collector.Clone()
			var imageURL string
			var publishedTime string
			var fullContent strings.Builder

			detailCollector.OnHTML("meta[property='og:image']", func(e *colly.HTMLElement) {
				imageURL = e.Attr("content")
			})
			detailCollector.OnHTML("meta[property='article:published_time']", func(e *colly.HTMLElement) {
				publishedTime = e.Attr("content")
			})

			// Extract article body content
			// CoinDesk uses various selectors, try multiple
			detailCollector.OnHTML("article p, .article-body p, [data-module-name='article-body'] p", func(e *colly.HTMLElement) {
				text := strings.TrimSpace(e.Text)
				if text != "" && len(text) > 20 { // Filter out short/empty paragraphs
					fullContent.WriteString(text)
					fullContent.WriteString("\n\n")
				}
			})

			if err := detailCollector.Visit(link); err != nil {
				return
			}

			now := time.Now().UTC()
			articles = append(articles, crawler.Article{
				Title:       title,
				SourceURL:   link,
				ImageURL:    imageURL,
				Content:     strings.TrimSpace(fullContent.String()), // Full article text
				Source:      "coindesk",
				CrawledAt:   now,
				PublishedAt: parsePublishedAt(publishedTime, now),
			})
		}
	})

	err := collector.Visit("https://www.coindesk.com/")
	if err != nil {
		return nil, crawler.ErrCrawlFailed
	}

	return articles, nil
}
