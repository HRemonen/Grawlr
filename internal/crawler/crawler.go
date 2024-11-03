/*
Package crawler provides a web crawler that uses a Fetcher to fetch web pages and Parsers to extract information.

The Crawler type can be used to crawl web pages up to a given depth and extract data using Parsers.

Example:

	f := fetcher.NewFetcher(&http.Client{
		Timeout: time.Second * 10,
	})
	p := []parser.Parser{}

	c := crawler.NewCrawler(f, p)

	err := c.Crawl("https://example.com/", 10)
	if err != nil {
		log.Fatal(err)
	}
*/
package crawler

import (
	"log"

	"github.com/HRemonen/Grawlr/internal/fetcher"
	"github.com/HRemonen/Grawlr/internal/parser"
)

// Crawler is a web crawler that uses a Fetcher to fetch web pages and Parsers to extract information.
type Crawler struct {
	Fetcher    *fetcher.Fetcher
	Parsers    []parser.Parser
	LinkParser *parser.LinkParser
}

// NewCrawler creates a new Crawler with the given Fetcher, Parsers, and LinkParser.
func NewCrawler(f *fetcher.Fetcher, p []parser.Parser) *Crawler {
	linkParser := parser.NewLinkParser()

	return &Crawler{
		Fetcher:    f,
		Parsers:    p,
		LinkParser: linkParser,
	}
}

// Crawl fetches the web page at the given URL and recursively crawls the pages
// linked from the page up to the given depth.
func (c *Crawler) Crawl(url string, depth int) error {
	return c.crawl(url, depth)
}

func (c *Crawler) crawl(url string, depth int) error {
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
