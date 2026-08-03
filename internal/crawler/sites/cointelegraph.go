package sites

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/0Hoag/cryptocheck-api/internal/crawler"
	"github.com/gocolly/colly/v2"
)

type coiTelegraphCrawler struct {
}

var rssImageSourcePattern = regexp.MustCompile(`(?i)\bsrc\s*=\s*["']([^"']+)["']`)

func NewCoinTelegraphCrawler() crawler.SiteCrawler {
	return &coiTelegraphCrawler{}
}

// newCoinTelegraphCollector returns a fresh collector for one crawl. Colly
// callbacks are registered on a collector instance, so sharing one across
// scheduled runs made OnXML handlers accumulate indefinitely.
func newCoinTelegraphCollector() *colly.Collector {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36"),
	)
	// Add Headers to look like a real browser
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Referer", "https://google.com")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
	})
	return c
}

// extractRSSImageURL handles feeds that embed their preview image in the HTML
// description instead of exposing a media:content element. The worker never
// fetches this URL; it is retained only for the browser/Telegram presentation.
func extractRSSImageURL(description string) string {
	matches := rssImageSourcePattern.FindStringSubmatch(description)
	if len(matches) != 2 {
		return ""
	}

	parsed, err := url.Parse(strings.TrimSpace(matches[1]))
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
		return ""
	}
	return parsed.String()
}

func (c *coiTelegraphCrawler) Name() string {
	return "cointelegraph"
}

func (c *coiTelegraphCrawler) Crawl(ctx context.Context) ([]crawler.Article, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	collector := newCoinTelegraphCollector()
	var articles []crawler.Article

	// Use RSS Feed
	collector.OnXML("//item", func(e *colly.XMLElement) {
		title := e.ChildText("title")
		link := e.ChildText("link")
		summary := e.ChildText("description") // Use description as summary

		// Images can be in media metadata or embedded in the RSS description.
		imageURL := e.ChildAttr("media:content", "url")
		if imageURL == "" {
			imageURL = e.ChildAttr("media:thumbnail", "url")
		}
		if imageURL == "" {
			imageURL = extractRSSImageURL(summary)
		}

		// Published Date
		pubDate := e.ChildText("pubDate")
		// Parse pubDate if needed, but for now we use time.Now() approximation or try parsing
		// Mon, 02 Jan 2006 15:04:05 MST
		publishedAt, err := time.Parse(time.RFC1123, pubDate)
		if err != nil {
			publishedAt = time.Now()
		}

		// Visit article page to extract full content
		detailCollector := collector.Clone()
		var fullContent strings.Builder

		detailCollector.OnHTML(".post-content p, .article__content p, .ct-prose p", func(e *colly.HTMLElement) {
			text := strings.TrimSpace(e.Text)
			if text != "" && len(text) > 20 {
				fullContent.WriteString(text)
				fullContent.WriteString("\n\n")
			}
		})

		// Visit the article page
		if link != "" {
			detailCollector.Visit(link)
		}

		if title != "" && link != "" {
			articles = append(articles, crawler.Article{
				Title:       strings.TrimSpace(title),
				Summary:     strings.TrimSpace(summary),
				SourceURL:   strings.TrimSpace(link),
				ImageURL:    imageURL,
				Content:     strings.TrimSpace(fullContent.String()), // Full article text
				Source:      "cointelegraph",
				CrawledAt:   time.Now(),
				PublishedAt: publishedAt,
			})
		}
	})

	err := collector.Visit("https://cointelegraph.com/rss")
	if err != nil {
		return nil, crawler.ErrCrawlFailed
	}

	collector.Wait()
	return articles, nil
}
