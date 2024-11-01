// Main file for the Grawlr application.
package main

import (
	"net/http"
	"time"

	"github.com/HRemonen/Grawlr/internal/crawler"
	"github.com/HRemonen/Grawlr/internal/fetcher"
)

func main() {
	f := fetcher.NewHTTPFetcher(&http.Client{
		Timeout: time.Second * 10,
	})
	c := crawler.NewHTTPCrawler(f)

	err := c.Crawl("https://web-scraping.dev/", 2)
	if err != nil {
		panic(err)
	}
}
