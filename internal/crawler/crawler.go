package crawler

import (
	"fmt"

	"github.com/HRemonen/Grawlr/internal/fetcher"
	"github.com/HRemonen/Grawlr/internal/parser"
)

// Crawler is an interface that defines the behavior of a web crawler.
type Crawler interface {
	Crawl(url string, depth int) error
}

// HttpCrawler is a web crawler that uses a Fetcher to fetch web pages.
type HttpCrawler struct {
	Fetcher fetcher.Fetcher
}

// NewHttpCrawler creates a new HttpCrawler with the given Fetcher.
func NewHttpCrawler(fetcher fetcher.Fetcher) *HttpCrawler {
	return &HttpCrawler{
		Fetcher: fetcher,
	}
}

// Crawl fetches the web page at the given URL and recursively crawls the pages
// linked from the page up to the given depth.
func (c *HttpCrawler) Crawl(url string, depth int) error {
	return c.crawl(url, depth)
}

func (c *HttpCrawler) crawl(url string, depth int) error {
	if depth == 0 {
		return nil
	}

	resp := c.Fetcher.Fetch(url)
	if resp.Error != nil {
		return resp.Error
	}

	links, err := parser.ExtractLinks(resp.Response.Body)
	fmt.Println("Links found on", url, ":", links)
	if err != nil {
		return err
	}

	fmt.Println("Crawling", url, "at depth", depth)
	for _, link := range links {
		if err := c.crawl(link, depth-1); err != nil {
			return err
		}
	}

	return nil
}
