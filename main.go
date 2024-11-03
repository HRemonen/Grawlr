// This file creates a new Fetcher and Crawler and starts the crawling process.
package main

import (
	"log"

	"github.com/HRemonen/Grawlr/internal/crawler"
	"github.com/HRemonen/Grawlr/internal/fetcher"
	"github.com/HRemonen/Grawlr/internal/parser"
	"github.com/HRemonen/Grawlr/internal/web"
)

func LoggingMiddleware() fetcher.Middleware {
	return func(req *web.Request) {
		log.Printf("Requesting URL: %s", req.URL.String())
	}
}

func main() {
	f := fetcher.NewFetcher(
		fetcher.WithMiddlewares(LoggingMiddleware()),
	)
	p := []parser.Parser{}
	c := crawler.NewCrawler(f, p)

	err := c.Crawl("https://hremonen.com", 10)
	if err != nil {
		panic(err)
	}
}
