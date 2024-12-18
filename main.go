// This file creates a new Fetcher and Crawler and starts the crawling process.
package main

import (
	"log"

	"github.com/HRemonen/Grawlr/grawl"
)

func main() {
	allowed := []string{
		"https://www.hremonen.com",
	}

	f := grawl.NewFetcher(
		grawl.WithAllowedURLs(allowed),
	)

	f.OnRequest(func(req *grawl.Request) {
		log.Println("[MAIN] - Visiting", req.URL.String())
	})

	f.OnScrape("a[href]", func(el *grawl.Element) {
		link := el.Attribute("href")

		log.Printf("[MAIN] - Found link %q -> %s", el.Text, link)

		absURL := el.Request.GetAbsoluteURL(link)

		err := f.Visit(absURL)
		if err != nil {
			log.Println("[MAIN] - ", err)
		}
	})

	err := f.Visit("https://www.hremonen.com")
	if err != nil {
		log.Println("[MAIN] - Error visiting start URL", err)
	}
}
