/*
Package fetcher provides a Fetcher that can be used to fetch web pages.

The Fetcher type uses an http.Client to fetch web pages and can be configured
with allowed and disallowed URLs as well as an option to ignore robots.txt.

Example:

	f := fetcher.NewFetcher(&http.Client{
		Timeout: time.Second * 10,
	}, fetcher.WithIgnoreRobots(true))

	resp, err := f.Request("https://example.com/")
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

var (
	// ErrForbiddenURL is returned when a URL is defined in the AllowedURLs setting.
	ErrForbiddenURL = errors.New("URL is forbidden")
	// ErrRobotsDisallowed is returned when a URL is disallowed by robots.txt.
	ErrRobotsDisallowed = errors.New("URL is disallowed by robots.txt")
)

// Options is a type for functional options that can be used to configure a Fetcher.
type Options func(f *Fetcher)

type Middleware func(req *web.Request)

// Fetcher is a Fetcher that uses an http.Client to fetch web pages.
type Fetcher struct {
	// Client is the http.Client used to fetch web pages.
	Client *http.Client
	// AllowedURLs is a list of URLs that are allowed to be fetched. Can be set with the WithAllowedURLs functional option.
	AllowedURLs []string
	// DisallowedURLs is a list of URLs that are disallowed to be fetched. Can be set with the WithDisallowedURLs functional option.
	DisallowedURLs []string
	// middlewares are an array of functions that are run before the request is made to the site.
	middlewares []Middleware
	// ignoreRobots is a flag that determines whether robots.txt should be ignored, defaults to false. Can be set with the WithIgnoreRobots functional option.
	ignoreRobots bool
	// robotsMap is a map of hostnames to robotstxt.RobotsData, which is used to cache robots.txt files.
	robotsMap map[string]*robotstxt.RobotsData
	// lock is a mutex used to synchronize access to the robotsMap.
	lock *sync.RWMutex
}

// NewFetcher creates a new Fetcher with the given http.Client.
func NewFetcher(options ...Options) *Fetcher {
	f := &Fetcher{
		Client:         http.DefaultClient,
		AllowedURLs:    []string{},
		DisallowedURLs: []string{},
		middlewares:    []Middleware{},
		ignoreRobots:   false,
		robotsMap:      make(map[string]*robotstxt.RobotsData),
		lock:           &sync.RWMutex{},
	}

	for _, option := range options {
		option(f)
	}

	return f
}

// WithClient is a functional option that sets the http.Client for the Fetcher.
func WithClient(client *http.Client) Options {
	return func(f *Fetcher) {
		f.Client = client
	}
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

// WithMiddleware is a functional option that adds a middleware to the Fetcher.
func WithMiddlewares(middlewares ...Middleware) Options {
	return func(f *Fetcher) {
		f.middlewares = append(f.middlewares, middlewares...)
	}
}

// Request requests the web page at the given URL if it is allowed to be fetched.
// It returns a web.Response with the response data or an error if the request fails.
func (f *Fetcher) Request(u string) (web.Response, error) {
	return f.scrape(u)
}

func (f *Fetcher) scrape(u string) (web.Response, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return web.Response{}, err
	}

	if err := f.checkRobots(parsedURL); err != nil {
		return web.Response{}, err
	}

	if err := f.checkFilters(parsedURL); err != nil {
		return web.Response{}, err
	}

	ctx := context.Background() // TODO: add functionality to cancel requests
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), http.NoBody)
	if err != nil {
		return web.Response{}, err
	}

	return f.fetch(req)
}

func (f *Fetcher) fetch(req *http.Request) (web.Response, error) {
	request := &web.Request{
		URL:     req.URL,
		Headers: &req.Header,
		Host:    req.URL.Host,
		Method:  req.Method,
		Body:    req.Body,
	}

	if err := f.handleMiddlewares(request); err != nil {
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
		Request:    request,
		Headers:    &resp.Header,
	}, nil
}

func (f *Fetcher) handleMiddlewares(req *web.Request) error {
	for _, m := range f.middlewares {
		m(req)
	}

	return nil
}

func (f *Fetcher) checkRobots(parsedURL *url.URL) error {
	if f.ignoreRobots {
		return nil
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

func (f *Fetcher) checkFilters(parsedURL *url.URL) error {
	u := parsedURL.String()

	if !f.isURLAllowed(u) {
		return ErrForbiddenURL
	}

	return nil
}

// isURLAllowed checks if the given URL is allowed to be fetched.
func (f *Fetcher) isURLAllowed(u string) bool {
	for _, disallowed := range f.DisallowedURLs {
		if u == disallowed {
			return false
		}
	}

	if len(f.AllowedURLs) == 0 {
		return true
	}

	for _, allowed := range f.AllowedURLs {
		if u == allowed {
			return true
		}
	}

	return false
}
