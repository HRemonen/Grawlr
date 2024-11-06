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
		log.Println("Visiting", req.URL.String())
	})

	f.OnResponse(func(res *grawl.Response) {
		log.Println("Visited", res.Request.URL.String())
		log.Println("Status code", res.StatusCode)
	})

	f.Visit("https://www.hremonen.com")
}
