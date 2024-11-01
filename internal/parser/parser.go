/*
Package parser provides a function to extract links from an HTML document.

The ExtractLinks function takes an io.ReadCloser and returns a slice of strings
containing the absolute URLs found in the document.

Example:

	links, err := parser.ExtractLinks(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
*/
package parser

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			return true, a.Val
		}
	}

	return false, ""
}

// ExtractLinks takes an io.ReadCloser and returns a slice of strings containing
// the absolute URLs found in the document.
func ExtractLinks(body io.Reader) ([]string, error) {
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
