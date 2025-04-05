package main

import (
	"log"

	"github.com/hremonen/grawlr"
)

func main() {
	allowed := []string{
		"https://www.hremonen.com",
	}

	f := grawlr.NewFetcher(
		grawlr.WithAllowedURLs(allowed),
	)

	f.RequestDo(func(req *grawlr.Request) {
		log.Println("[MAIN] - Visiting", req.URL.String())
	})

	f.HtmlDo("a[href]", func(el *grawlr.Element) {
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
