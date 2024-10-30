package fetcher

import "net/http"

type Fetcher interface {
	Fetch(url string) Response
}

type Response struct {
	Error    error
	Response *http.Response
}

type HttpFetcher struct {
	Client *http.Client
}

func NewHttpFetcher(client *http.Client) *HttpFetcher {
	return &HttpFetcher{
		Client: client,
	}
}

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
