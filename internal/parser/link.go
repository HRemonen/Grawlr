package parser

import (
	"strings"

	"github.com/HRemonen/Grawlr/internal/web"
	"golang.org/x/net/html"
)

type LinkParser struct {
	Response web.Response
}

// NewLinkParser creates a new LinkParser with the given Response.
func NewLinkParser(r web.Response) *LinkParser {
	return &LinkParser{
		Response: r,
	}
}

// ExtractLinks takes an io.Reader and returns a slice of strings containing
// the absolute URLs found in the document.
func (p *LinkParser) Parse() ([]string, error) {
	links := []string{}
	tokenizer := html.NewTokenizer(p.Response.Body)

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
