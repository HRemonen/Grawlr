/*
Package web provides types for internal Response objects.
*/
package web

import (
	"context"
	"io"
	"net/http"
)

// Response is a representation of the response from a Fetcher.
type Response struct {
	StatusCode int
	Body       io.Reader
	Ctx        *context.Context
	Request    *http.Request
	Headers    *http.Header
}
