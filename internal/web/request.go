package web

import (
	"io"
	"net/http"
	"net/url"
)

type Request struct {
	URL     *url.URL
	Headers *http.Header
	Host    string
	Method  string
	Body    io.Reader
}
