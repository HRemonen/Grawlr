package main

import (
	"log"

	grawlr "github.com/HRemonen/Grawlr"
)

func main() {
	allowed := []string{
		"https://www.hremonen.com",
	}

	h := grawlr.NewHarvester(
		grawlr.WithAllowedURLs(allowed),
		grawlr.WithDepthLimit(1), // Set the maximum depth to 1
	)

	h.RequestDo(func(req *grawlr.Request) {
		log.Println("[MAIN] - Visiting", req.URL.String())
	})

	count := 0
	h.ResponseDo(func(res *grawlr.Response) {
		count++
	})

	h.HtmlDo("a[href]", func(el *grawlr.HtmlElement) {
		link := el.Attribute("href")

		log.Printf("[MAIN] - Found link %q -> %s", el.Text, link)

		absURL := el.Request.GetAbsoluteURL(link)

		err := el.Request.Visit(absURL) // Use el.Request.Visit to preserve the depth context
		if err != nil {
			log.Println("[MAIN] - ", err)
		}
	})

	err := h.Visit("https://www.hremonen.com")
	if err != nil {
		log.Println("[MAIN] - Error visiting start URL", err)
	}

	log.Println("[MAIN] - Visited", count, "pages")
}
