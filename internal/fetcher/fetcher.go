/*
Package fetcher provides a Fetcher that can be used to fetch web pages.

The Fetcher type uses an http.Client to fetch web pages and can be configured
with allowed and disallowed URLs as well as an option to ignore robots.txt.

Example:

	f := fetcher.NewFetcher(&http.Client{
		Timeout: time.Second * 10,
	}, fetcher.WithIgnoreRobots(true))

	resp, err := f.Fetch("https://example.com/")
	if err != nil {
		log.Fatal(err)
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

// Options is a type for functional options that can be used to configure a Fetcher.
type Options func(f *Fetcher)

// Fetcher is a Fetcher that uses an http.Client to fetch web pages.
type Fetcher struct {
	// Client is the http.Client used to fetch web pages.
	Client *http.Client
	// AllowedURLs is a list of URLs that are allowed to be fetched.
	AllowedURLs []string
	// DisallowedURLs is a list of URLs that are disallowed to be fetched.
	DisallowedURLs []string
	// ignoreRobots is a flag that determines whether robots.txt should be ignored.
	ignoreRobots bool
	// robotsMap is a map of hostnames to robotstxt.RobotsData, which is used to cache robots.txt files.
	robotsMap map[string]*robotstxt.RobotsData
	// lock is a mutex used to synchronize access to the robotsMap.
	lock *sync.RWMutex
}

// NewFetcher creates a new Fetcher with the given http.Client.
func NewFetcher(client *http.Client, options ...Options) *Fetcher {
	f := &Fetcher{
		Client:       client,
		ignoreRobots: false,
		robotsMap:    make(map[string]*robotstxt.RobotsData),
		lock:         &sync.RWMutex{},
	}

	for _, option := range options {
		option(f)
	}

	return f
}

// WithAllowedURLs is a functional option that sets the allowed URLs for the Fetcher.
func WithAllowedURLs(urls []string) Options {
	return func(f *Fetcher) {
		f.AllowedURLs = urls
	}
}

// WithDisallowedURLs is a functional option that sets the disallowed URLs for the Fetcher.
func WithDisallowedURLs(urls []string) Options {
	return func(f *Fetcher) {
		f.DisallowedURLs = urls
	}
}

// WithIgnoreRobots is a functional option that sets the ignoreRobots flag for the Fetcher.
func WithIgnoreRobots(ignore bool) Options {
	return func(f *Fetcher) {
		f.ignoreRobots = ignore
	}
}

// Fetch fetches the web page at the given URL and return a custom Response object.
func (f *Fetcher) Fetch(u string) (web.Response, error) {
	ctx := context.Background()

	if err := f.checkRobots(u); err != nil {
		return web.Response{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
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

func (f *Fetcher) checkRobots(u string) error {
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
		resp, err := f.Client.Get(robotURL) //nolint: noctx // we don't need a context here
		if err != nil {
			return err
		}

		defer resp.Body.Close() //nolint: errcheck // because we don't care about the error here

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
