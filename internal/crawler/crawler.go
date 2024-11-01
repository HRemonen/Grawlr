/*
Package crawler provides a web crawler that can be used to crawl web pages.

The Crawler interface defines the behavior of a web crawler. The HTTPCrawler
type implements the Crawler interface using a Fetcher to fetch web pages.

Example:

	f := fetcher.NewHTTPFetcher(&http.Client{
		Timeout: time.Second * 10,
	})
	c := crawler.NewHTTPCrawler(f)

	err := c.Crawl("https://example.com/", 2)
	if err != nil {
		panic(err)
	}
*/
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

// HTTPCrawler is a web crawler that uses a Fetcher to fetch web pages.
type HTTPCrawler struct {
	Fetcher fetcher.Fetcher
}

// NewHTTPCrawler creates a new HTTPCrawler with the given Fetcher.
func NewHTTPCrawler(f fetcher.Fetcher) *HTTPCrawler {
	return &HTTPCrawler{
		Fetcher: f,
	}
}

// Crawl fetches the web page at the given URL and recursively crawls the pages
// linked from the page up to the given depth.
func (c *HTTPCrawler) Crawl(url string, depth int) error {
	return c.crawl(url, depth)
}

func (c *HTTPCrawler) crawl(url string, depth int) error {
	if depth == 0 {
		return nil
	}

	fr := c.Fetcher.Fetch(url)
	if fr.Error != nil {
		return fr.Error
	}

	links, err := parser.ExtractLinks(fr.Response.Body)
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
