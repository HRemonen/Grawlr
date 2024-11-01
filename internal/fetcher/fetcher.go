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
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/HRemonen/Grawlr/internal/web"

	"github.com/temoto/robotstxt"
)

// ErrRobotsDisallowed is returned when a URL is disallowed by robots.txt.
var ErrRobotsDisallowed = errors.New("URL is disallowed by robots.txt")

// Fetcher is an interface that defines the behavior of a web page fetcher.
type Fetcher interface {
	Fetch(url string) (web.Response, error)
}

// HTTPFetcher is a Fetcher that uses an http.Client to fetch web pages.
type HTTPFetcher struct {
	Client       *http.Client
	ignoreRobots bool
	robotsMap    map[string]*robotstxt.RobotsData
	lock         *sync.Mutex
}

// NewHTTPFetcher creates a new HTTPFetcher with the given http.Client.
func NewHTTPFetcher(client *http.Client) *HTTPFetcher {
	return &HTTPFetcher{
		Client:       client,
		ignoreRobots: false,
		robotsMap:    make(map[string]*robotstxt.RobotsData),
		lock:         &sync.Mutex{},
	}
}

// Fetch fetches the web page at the given URL and return a custom Response object.
func (f *HTTPFetcher) Fetch(url string) (web.Response, error) {
	ctx := context.Background()

	if err := f.checkRobots(url); err != nil {
		return web.Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return web.Response{}, err
	}

	resp, err := f.Client.Do(req)
	if err != nil {
		return web.Response{}, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error closing response body: %v for request of: %v", err, req.URL)
		}
	}()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return web.Response{}, err
	}

	body := bytes.NewReader(b)

	return web.Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Ctx:        &ctx,
		Request:    req,
		Headers:    &resp.Header,
	}, nil
}

func (f *HTTPFetcher) checkRobots(u string) error {
	if f.ignoreRobots {
		return nil
	}

	parsedURL, err := url.Parse(u)
	if err != nil {
		return err
	}

	f.lock.Lock()
	robot, ok := f.robotsMap[parsedURL.Host]
	f.lock.Unlock()

	if !ok {
		robotURL := parsedURL.Scheme + "://" + parsedURL.Host + "/robots.txt"
		resp, err := f.Client.Get(robotURL)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		robot, err = robotstxt.FromResponse(resp)
		if err != nil {
			return err
		}

		f.lock.Lock()
		f.robotsMap[parsedURL.Host] = robot
		f.lock.Unlock()
	}

	if !robot.TestAgent(parsedURL.Path, "Grawlr") {
		return ErrRobotsDisallowed
	}

	return nil
}
