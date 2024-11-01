/*
Package fetcher provides a simple interface for fetching web pages.

The Fetcher interface defines the behavior of a web page fetcher. The HttpFetcher
type implements the Fetcher interface using an http.Client to fetch web pages.

Example:

	f := fetcher.NewHTTPFetcher(&http.Client{
		Timeout: time.Second * 10,
	})
	resp := f.Fetch("https://example.com/")
	if resp.Error != nil {
		log.Fatal(resp.Error)
	}
*/
package fetcher

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// Fetcher is an interface that defines the behavior of a web page fetcher.
type Fetcher interface {
	Fetch(url string) (Response, error)
}

// Response is a representation of the response from a Fetcher.
type Response struct {
	StatusCode int
	Body       io.Reader
	Ctx        *context.Context
	Request    *http.Request
	Headers    *http.Header
}

// HTTPFetcher is a Fetcher that uses an http.Client to fetch web pages.
type HTTPFetcher struct {
	Client *http.Client
}

// NewHTTPFetcher creates a new HTTPFetcher with the given http.Client.
func NewHTTPFetcher(client *http.Client) *HTTPFetcher {
	return &HTTPFetcher{
		Client: client,
	}
}

// Fetch fetches the web page at the given URL and return a custom Response object.
func (f *HTTPFetcher) Fetch(url string) (Response, error) {
	ctx := context.Background()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return Response{}, err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return Response{}, err
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, err
	}

	body := bytes.NewReader(b)

	return Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Ctx:        &ctx,
		Request:    req,
		Headers:    &resp.Header,
	}, nil
}
