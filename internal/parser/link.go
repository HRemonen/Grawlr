package parser

import (
	"strings"

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

	for {
		tt := tokenizer.Next()

		switch {
		case tt == html.ErrorToken: // End of the document
			return links, nil
		case tt == html.StartTagToken:
			t := tokenizer.Token()

			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			ok, href := getHref(t)
			if !ok {
				continue
			}

			// Checking that the href is absolute
			// TODO: Handle relative URLs
			hasProto := strings.Index(href, "http") == 0
			if hasProto {
				links = append(links, href)
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
