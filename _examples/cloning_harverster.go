package main

import (
	"log"

	grawlr "github.com/HRemonen/Grawlr"
)

func main() {
	// Initial configuration for the original Harvester
	h1 := grawlr.NewHarvester(
		grawlr.WithAllowedURLs([]string{"https://www.hremonen.com"}),
		grawlr.WithDepthLimit(2),
	)

	h1.RequestDo(func(req *grawlr.Request) {
		log.Println("[H1] - Visiting", req.URL.String())
	})

	h1.HtmlDo("a[href]", func(el *grawlr.HtmlElement) {
		link := el.Attribute("href")
		log.Printf("[H1] - Found link %q -> %s", el.Text, link)
	})

	// Clone the Harvester
	h2 := h1.Clone()

	// Add additional behavior to the cloned Harvester
	h2.RequestDo(func(req *grawlr.Request) {
		log.Println("[H2] - Visiting", req.URL.String())
	})

	h2.HtmlDo("img[src]", func(el *grawlr.HtmlElement) {
		src := el.Attribute("src")
		log.Printf("[H2] - Found image source -> %s", src)
	})

	// Use the original Harvester
	err := h1.Visit("https://www.hremonen.com")
	if err != nil {
		log.Println("[H1] - Error visiting start URL:", err)
	}

	// Use the cloned Harvester
	err = h2.Visit("https://www.hremonen.com")
	if err != nil {
		log.Println("[H2] - Error visiting start URL:", err)
	}
}
