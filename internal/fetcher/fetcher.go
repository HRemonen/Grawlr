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

// HttpFetcher is a Fetcher that uses an http.Client to fetch web pages.
type HttpFetcher struct {
	Client *http.Client
}

// NewHttpFetcher creates a new HttpFetcher with the given http.Client.
func NewHttpFetcher(client *http.Client) *HttpFetcher {
	return &HttpFetcher{
		Client: client,
	}
}

// Fetch fetches the web page at the given URL and return a custom Response object.
func (f *HttpFetcher) Fetch(url string) Response {
	resp, err := f.Client.Get(url)
	if err != nil {
		return Response{
			Response: nil,
			Error:    err,
		}
	}

	return Response{
		Response: resp,
		Error:    nil,
	}
}
