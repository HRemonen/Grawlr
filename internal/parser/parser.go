package parser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func getHref(t html.Token) (bool, string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return true, a.Val
		}
	}

	return false, ""
}

func ExtractLinks(body io.ReadCloser) ([]string, error) {
	defer body.Close()

	links := []string{}
	tokenizer := html.NewTokenizer(body)

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
