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

// Parser is an interface that defines the behavior of a parser.
type Parser interface {
	Parse() ([]string, error)
}
