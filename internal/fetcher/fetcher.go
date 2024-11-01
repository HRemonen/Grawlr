/*
Package fetcher provides a simple interface for fetching web pages.

The Fetcher interface defines the behavior of a web page fetcher. The HttpFetcher
type implements the Fetcher interface using an http.Client to fetch web pages.

Example:

	f := fetcher.NewHTTPFetcher(&http.Client{
		Timeout: time.Second * 10,
	})
	resp := fetcher.Fetch("https://example.com/")
	if resp.Error != nil {
		log.Fatal(resp.Error)
	}
*/
package fetcher

import "net/http"

// Fetcher is an interface that defines the behavior of a web page fetcher.
type Fetcher interface {
	Fetch(url string) Response
}

// Response is a custom response object that contains an error and an http.Response.
type Response struct {
	Error    error
	Response *http.Response
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
func (f *HTTPFetcher) Fetch(url string) Response {
	resp, err := f.Client.Get(url)
	if err != nil {
		return Response{
			Response: nil,
			Error:    err,
		}
	}

	defer resp.Body.Close()
	return Response{
		Response: resp,
		Error:    nil,
	}
}
