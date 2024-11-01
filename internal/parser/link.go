package parser

import (
	"fmt"
	"log"
	"net/url"

	"github.com/HRemonen/Grawlr/internal/web"
	"golang.org/x/net/html"
)

// LinkParser is a parser that extracts links from a web.Response.
type LinkParser struct{}

// NewLinkParser creates a new LinkParser with the given Response.
func NewLinkParser() *LinkParser {
	return &LinkParser{}
}

// Parse returns a slice of strings containing
// the absolute URLs found in the document.
func (p *LinkParser) Parse(content web.Response) ([]string, error) {
	links := []string{}
	tokenizer := html.NewTokenizer(content.Body)

	// Ensure content.Request.URL is a valid URL
	baseURL, err := url.Parse(content.Request.URL.String())
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %v", err)
	}

	for {
		tt := tokenizer.Next()

		switch {
		case tt == html.ErrorToken: // End of the document
			return links, nil
		case tt == html.StartTagToken:
			t := tokenizer.Token()

			if t.Data != "a" {
				continue
			}

			ok, href := getHref(t)
			if !ok {
				continue
			}

			linkURL, err := baseURL.Parse(href)
			if err != nil {
				log.Println("Error parsing link:", href, err)
				continue
			}

			// Check if the scheme is valid
			if linkURL.Scheme == "http" || linkURL.Scheme == "https" {
				links = append(links, linkURL.String())
			}
		}
	}
}

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return true, a.Val
		}
	}

	return false, ""
}
