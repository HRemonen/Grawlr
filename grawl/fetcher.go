/*
Copyright 2024 Henri Remonen

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package grawl

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/PuerkitoBio/goquery"
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

// ReqMiddleware is a type for request middlewares that can be used to modify a Request before it is fetched.
type ReqMiddleware func(req *Request)

// ResMiddleware is a type for response middlewares that can be used to modify a Response after it is fetched.
type ResMiddleware func(res *Response)

// Fetcher is a Fetcher that uses an http.Client to fetch web pages.
type Fetcher struct {
	// Client is the http.Client used to fetch web pages.
	Client *http.Client
	// AllowedURLs is a list of URLs that are allowed to be fetched. Can be set with the WithAllowedURLs functional option.
	AllowedURLs []string
	// DisallowedURLs is a list of URLs that are disallowed to be fetched. Can be set with the WithDisallowedURLs functional option.
	DisallowedURLs []string
	// requestMiddlewares is a list of request middlewares that are applied to each request. Can be set with the OnRequest functional option.
	requestMiddlewares []ReqMiddleware
	// responseMiddlewares is a list of response middlewares that are applied to each response. Can be set with the OnResponse functional option.
	responseMiddlewares []ResMiddleware
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
		Client:              http.DefaultClient,
		AllowedURLs:         []string{},
		DisallowedURLs:      []string{},
		requestMiddlewares:  make([]ReqMiddleware, 0, 4),
		responseMiddlewares: make([]ResMiddleware, 0, 4),
		ignoreRobots:        false,
		robotsMap:           make(map[string]*robotstxt.RobotsData),
		lock:                &sync.RWMutex{},
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

// OnRequest is a functional option that adds a request middleware to the Fetcher.
func (f *Fetcher) OnRequest(middleware ReqMiddleware) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.requestMiddlewares = append(f.requestMiddlewares, middleware)
}

// OnResponse is a functional option that adds a response middleware to the Fetcher.
func (f *Fetcher) OnResponse(middleware ResMiddleware) {
	f.lock.Lock()
	defer f.lock.Unlock()

	f.responseMiddlewares = append(f.responseMiddlewares, middleware)
}

// Visit requests the web page at the given URL if it is allowed to be fetched.
// It returns a Response with the response data or an error if the request fails.
func (f *Fetcher) Visit(u string) error {
	return f.visit(u)
}

func (f *Fetcher) visit(u string) error {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return err
	}

	if err := f.checkRobots(parsedURL); err != nil {
		return err
	}

	if err := f.checkFilters(parsedURL); err != nil {
		return err
	}

	ctx := context.Background() // TODO: add functionality to cancel requests
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), http.NoBody)
	if err != nil {
		return err
	}

	return f.fetch(req)
}

func (f *Fetcher) fetch(req *http.Request) error {
	request := &Request{
		URL:     req.URL,
		Headers: &req.Header,
		Host:    req.URL.Host,
		Method:  req.Method,
		Body:    req.Body,
	}

	f.handleOnRequest(request)

	res, err := f.Client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err := res.Body.Close(); err != nil {
			log.Printf("error closing response body: %v for request of: %v", err, req.URL)
		}
	}()

	// Read the full response body into `b`.
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Create a new reader from `b` for repeated reads.
	body := bytes.NewReader(b)

	// Parse the body into a goquery document (this consumes the body reader).
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return err
	}

	// Reset the body reader for later use in `OnResponse`.
	_, err = body.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	response := &Response{
		StatusCode: res.StatusCode,
		Headers:    &res.Header,
		Request:    request,
		Body:       body, // Assign the reusable reader here.
		Document:   doc,
	}

	// Pass the response to the `OnResponse` handler.
	f.handleOnResponse(response)

	return nil
}

func (f *Fetcher) handleOnRequest(req *Request) {
	for _, m := range f.requestMiddlewares {
		m(req)
	}
}

func (f *Fetcher) handleOnResponse(res *Response) {
	for _, m := range f.responseMiddlewares {
		m(res)
	}
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
		res, err := f.Client.Get(robotURL) //nolint: noctx // we don't need a context here
		if err != nil {
			return err
		}

		defer res.Body.Close() //nolint: errcheck // because we don't care about the error here

		robot, err = robotstxt.FromResponse(res)
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
