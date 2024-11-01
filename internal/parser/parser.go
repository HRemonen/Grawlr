/*
Package parser provides a simple interface for parsing web pages.
*/
package parser

import "github.com/HRemonen/Grawlr/internal/web"

// Parser is an interface that defines the behavior of a parser.
type Parser interface {
	Parse(content web.Response) ([]string, error)
}
