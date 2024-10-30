package main

import (
	"net/http"

	"github.com/HRemonen/Grawlr/internal/crawler"
	"github.com/HRemonen/Grawlr/internal/fetcher"
)

func main() {
	client := &http.Client{}
	fetcher := fetcher.NewHttpFetcher(client)
	crawler := crawler.NewHttpCrawler(fetcher)
	crawler.Crawl("https://web-scraping.dev/", 2)
}
