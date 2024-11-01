// Main file for the Grawlr application.
package main

import (
	"net/http"
	"time"

	"github.com/HRemonen/Grawlr/internal/crawler"
	"github.com/HRemonen/Grawlr/internal/fetcher"
	"github.com/HRemonen/Grawlr/internal/parser"
)

func main() {
	f := fetcher.NewHTTPFetcher(&http.Client{
		Timeout: time.Second * 10,
	})
	p := []parser.Parser{}
	c := crawler.NewHTTPCrawler(f, p)

	err := c.Crawl("https://hremonen.com", 10)
	if err != nil {
		panic(err)
	}
}
