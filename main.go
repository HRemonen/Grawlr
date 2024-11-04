// This file creates a new Fetcher and Crawler and starts the crawling process.
package main

import (
	"log"

	"github.com/HRemonen/Grawlr/grawl"
)

func LoggingMiddleware() grawl.Middleware {
	return func(req *grawl.Request) {
		log.Printf("Requesting URL: %s", req.URL.String())
	}
}

func main() {

}
