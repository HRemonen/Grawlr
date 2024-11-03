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
	"log"

	"github.com/HRemonen/Grawlr/internal/fetcher"
	"github.com/HRemonen/Grawlr/internal/parser"
)

// Crawler is an interface that defines the behavior of a web crawler.
type Crawler interface {
	Crawl(url string, depth int) error
}

// HTTPCrawler is a web crawler that uses a Fetcher to fetch web pages and Parsers to extract information.
type HTTPCrawler struct {
	Fetcher    *fetcher.Fetcher
	Parsers    []parser.Parser
	LinkParser *parser.LinkParser
}

// NewHTTPCrawler creates a new HTTPCrawler with the given Fetcher, Parsers, and LinkParser.
func NewHTTPCrawler(f *fetcher.Fetcher, p []parser.Parser) *HTTPCrawler {
	linkParser := parser.NewLinkParser()

	return &HTTPCrawler{
		Fetcher:    f,
		Parsers:    p,
		LinkParser: linkParser,
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

	// Fetch the page content
	res, err := c.Fetcher.Fetch(url)
	if err != nil {
		return err
	}

	log.Println("Crawling", url, "at depth", depth)

	// Use non-link parsers to extract data from the page
	for _, parser := range c.Parsers {
		parsedData, err := parser.Parse(res)
		if err != nil {
			log.Printf("Error parsing %s with parser %T: %v", url, parser, err)
			continue
		}
		log.Printf("Data found by %T on %s: %v", parser, url, parsedData)
	}

	// Use the dedicated LinkParser to extract links for further crawling
	links, err := c.LinkParser.Parse(res)
	if err != nil {
		log.Printf("Error parsing links on %s: %v", url, err)
	}
	log.Println("Links found on", url, ":", links)

	// Recursively crawl each discovered link
	for _, link := range links {
		if err := c.crawl(link, depth-1); err != nil {
			log.Printf("Error crawling %s: %v", link, err)
		}
	}

	return nil
}
