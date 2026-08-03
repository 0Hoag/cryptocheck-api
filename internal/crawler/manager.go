package crawler

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/0Hoag/cryptocheck-api/pkg/log"
)

// Manager handles multiple crawlers
type Manager struct {
	crawlers []SiteCrawler
	l        log.Logger
}

func NewManager(l log.Logger) *Manager {
	return &Manager{
		crawlers: make([]SiteCrawler, 0),
		l:        l,
	}
}

func (m *Manager) Register(c SiteCrawler) {
	m.crawlers = append(m.crawlers, c)
}

func (m *Manager) Run(ctx context.Context) ([]Article, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allArticles []Article
	var successfulCrawlers int

	for _, c := range m.crawlers {
		wg.Add(1)
		go func(crawler SiteCrawler) {
			defer wg.Done()

			// Set timeout for each crawler
			ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			defer cancel()

			articles, err := crawler.Crawl(ctx)
			if err != nil {
				m.l.Errorf(ctx, "Crawler %s failed: %v", crawler.Name(), err)
				return
			}
			m.l.Infof(ctx, "Crawler %s fetched %d articles", crawler.Name(), len(articles))

			mu.Lock()
			successfulCrawlers++
			allArticles = append(allArticles, articles...)
			mu.Unlock()
		}(c)
	}

	wg.Wait()
	if len(m.crawlers) > 0 && successfulCrawlers == 0 {
		return nil, fmt.Errorf("all %d registered crawlers failed", len(m.crawlers))
	}
	return normalizeArticles(allArticles), nil
}

// normalizeArticles makes concurrent crawler output deterministic before the
// worker chooses a batch. Newest articles win and an article URL is published
// at most once even if multiple sources surface the same link.
func normalizeArticles(articles []Article) []Article {
	unique := make(map[string]struct{}, len(articles))
	result := make([]Article, 0, len(articles))
	for _, article := range articles {
		if article.SourceURL != "" {
			if _, seen := unique[article.SourceURL]; seen {
				continue
			}
			unique[article.SourceURL] = struct{}{}
		}
		result = append(result, article)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].PublishedAt.Equal(result[j].PublishedAt) {
			return result[i].SourceURL < result[j].SourceURL
		}
		return result[i].PublishedAt.After(result[j].PublishedAt)
	})
	return result
}
