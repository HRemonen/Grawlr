package main

import (
	"log"

	grawlr "github.com/HRemonen/Grawlr"
	"github.com/hremonen/grawlr"
)

func main() {
	allowed := []string{
		"https://www.hremonen.com",
	}

	h := grawlr.NewHarvester(
		grawlr.WithAllowedURLs(allowed),
	)

	h.RequestDo(func(req *grawlr.Request) {
		log.Println("[MAIN] - Visiting", req.URL.String())
	})

	h.HtmlDo("a[href]", func(el *grawlr.HtmlElement) {
		link := el.Attribute("href")

		log.Printf("[MAIN] - Found link %q -> %s", el.Text, link)

		absURL := el.Request.GetAbsoluteURL(link)

		err := h.Visit(absURL)
		if err != nil {
			log.Println("[MAIN] - ", err)
		}
	})

	err := h.Visit("https://www.hremonen.com")
	if err != nil {
		log.Println("[MAIN] - Error visiting start URL", err)
	}
}
