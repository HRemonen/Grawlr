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
package grawlr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/temoto/robotstxt"
)

var (
	// ErrForbiddenURL is returned when a URL is defined in the AllowedURLs setting.
	ErrForbiddenURL = func(u string) error {
		return fmt.Errorf("URL %s is forbidden", u)
	}
	// ErrRobotsDisallowed is returned when a URL is disallowed by robots.txt.
	ErrRobotsDisallowed = func(u string) error {
		return fmt.Errorf("URL %s is disallowed by robots.txt", u)
	}
	// ErrVisitedURL is returned when a URL has already been visited.
	ErrVisitedURL = func(u string) error {
		return fmt.Errorf("URL %s has already been visited", u)
	}
	// ErrDepthLimitExceeded is returned when the maximum depth limit is exceeded.
	ErrDepthLimitExceeded = func(depth, limit int) error {
		return fmt.Errorf("depth limit exceeded: %d > %d", depth, limit)
	}
)

// Options is a type for functional options that can be used to configure a Harvester.
type Options func(h *Harvester)

// ReqMiddleware is a type for request middlewares that can be used to modify a Request before it is fetched.
type ReqMiddleware func(req *Request)

// ResMiddleware is a type for response middlewares that can be used to modify a Response after it is fetched.
type ResMiddleware func(res *Response)

type (
	HtmlCallback   func(el *HtmlElement)
	HtmlMiddleware struct {
		Selector string
		Function HtmlCallback
	}
)

// Harvester is a Harvester that uses an http.Client to fetch web pages.
type Harvester struct {
	// Client is the http.Client used to fetch web pages.
	Client *http.Client
	// AllowedURLs is a list of URLs that are allowed to be fetched. Can be set with the WithAllowedURLs functional option.
	AllowedURLs []string
	// DisallowedURLs is a list of URLs that are disallowed to be fetched. Can be set with the WithDisallowedURLs functional option.
	DisallowedURLs []string
	// DepthLimit is the maximum depth of links to follow. If set to 0, all links are followed. Can be set with the WithDepthLimit functional option.
	DepthLimit int
	// AllowRevisit is a flag that determines whether to allow revisiting URLs. If set to true, URLs can be revisited even if they have already been visited. Defaults to false.
	AllowRevisit bool
	// Context is the context used to optionally cancel ALL harvester's requests. Can be set with the WithContext functional option.
	Context context.Context
	// store is a Storer that is used to cache visited URLs.
	store Storer
	// requestMiddlewares is a list of request middlewares that are applied to each request. Can be set with the RequestDo functional option.
	requestMiddlewares []ReqMiddleware
	// responseMiddlewares is a list of response middlewares that are applied to each response. Can be set with the ResponseDo functional option.
	responseMiddlewares []ResMiddleware
	// htmlMiddlewares is a list of scrape middlewares that are applied to each Html HtmlElement. Can be set with the HtmlDo functional option.
	htmlMiddlewares []HtmlMiddleware
	// ignoreRobots is a flag that determines whether robots.txt should be ignored, defaults to false. Can be set with the WithIgnoreRobots functional option.
	ignoreRobots bool
	// robotsMap is a map of hostnames to robotstxt.RobotsData, which is used to cache robots.txt files.
	robotsMap map[string]*robotstxt.RobotsData
	// mu is a mutex used to synchronize access to the robotsMap.
	mu sync.RWMutex
}

// NewHarvester creates a new Harvester with the given http.Client.
func NewHarvester(options ...Options) *Harvester {
	h := &Harvester{
		Client:              http.DefaultClient,
		AllowedURLs:         []string{},
		DisallowedURLs:      []string{},
		DepthLimit:          0,
		AllowRevisit:        false,
		Context:             context.Background(),
		store:               NewInMemoryStore(),
		requestMiddlewares:  make([]ReqMiddleware, 0, 4),
		responseMiddlewares: make([]ResMiddleware, 0, 4),
		htmlMiddlewares:     make([]HtmlMiddleware, 0, 4),
		ignoreRobots:        false,
		robotsMap:           make(map[string]*robotstxt.RobotsData),
		mu:                  sync.RWMutex{},
	}

	for _, option := range options {
		option(h)
	}

	return h
}

// Clone returns a new Harvester with the same options as the original
// except for the middleware functions.
func (h *Harvester) Clone() *Harvester {
	// Create a new Harvester with the same options as the original
	clone := &Harvester{
		Client:              h.Client,
		AllowedURLs:         h.AllowedURLs,
		DisallowedURLs:      h.DisallowedURLs,
		DepthLimit:          h.DepthLimit,
		AllowRevisit:        h.AllowRevisit,
		Context:             h.Context,
		store:               h.store,
		requestMiddlewares:  make([]ReqMiddleware, 0, 4),
		responseMiddlewares: make([]ResMiddleware, 0, 4),
		htmlMiddlewares:     make([]HtmlMiddleware, 0, 4),
		ignoreRobots:        h.ignoreRobots,
		robotsMap:           h.robotsMap,
		mu:                  sync.RWMutex{},
	}

	return clone
}

// WithClient is a functional option that sets the http.Client for the Harvester.
func WithClient(client *http.Client) Options {
	return func(h *Harvester) {
		h.Client = client
	}
}

// WithAllowRevisit is a functional option that sets the AllowRevisit flag for the Harvester.
func WithAllowRevisit(allow bool) Options {
	return func(h *Harvester) {
		h.AllowRevisit = allow
	}
}

// WithAllowedURLs is a functional option that sets the allowed URLs for the Harvester.
func WithAllowedURLs(urls []string) Options {
	return func(h *Harvester) {
		h.AllowedURLs = urls
	}
}

// WithDisallowedURLs is a functional option that sets the disallowed URLs for the Harvester.
func WithDisallowedURLs(urls []string) Options {
	return func(h *Harvester) {
		h.DisallowedURLs = urls
	}
}

// WithDepthLimit is a functional option that sets the maximum depth for the Harvester.
func WithDepthLimit(depth int) Options {
	return func(h *Harvester) {
		h.DepthLimit = depth
	}
}

// WithContext is a functional option that sets the context for the Harvester.
func WithContext(ctx context.Context) Options {
	return func(h *Harvester) {
		h.Context = ctx
	}
}

// WithStore is a functional option that sets the Storer for the Harvester.
// See the Storer interface in store.go for more information.
func WithStore(store Storer) Options {
	return func(h *Harvester) {
		h.store = store
	}
}

// WithIgnoreRobots is a functional option that sets the ignoreRobots flag for the Harvester.
func WithIgnoreRobots(ignore bool) Options {
	return func(h *Harvester) {
		h.ignoreRobots = ignore
	}
}

// RequestDo is a functional option that adds a request middleware to the Harvester.
// Triggers the given ReqMiddleware for each request before it is fetched.
func (h *Harvester) RequestDo(mw ReqMiddleware) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.requestMiddlewares = append(h.requestMiddlewares, mw)
}

// ResponseDo is a functional option that adds a response middleware to the Harvester.
// Triggers the given ResMiddleware for each response after a request.
func (h *Harvester) ResponseDo(mw ResMiddleware) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.responseMiddlewares = append(h.responseMiddlewares, mw)
}

// HtmlDo is a functional option that adds a Html middleware to the Harvester.
// HtmlCallback is a function that is executed on every Html HtmlElement that matches the given GoQuery selector.
//
// SEE GoQuery documentation for more information on selectors: https://pkg.go.dev/github.com/PuerkitoBio/goquery
func (h *Harvester) HtmlDo(gqSelector string, fn HtmlCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.htmlMiddlewares = append(h.htmlMiddlewares, HtmlMiddleware{
		Selector: gqSelector,
		Function: fn,
	})
}

// Visit requests the web page at the given URL if it is allowed to be fetched.
// It returns a Response with the response data or an error if the request fails.
func (h *Harvester) Visit(u string) error {
	return h.fetch(u, http.MethodGet, 0)
}

func (h *Harvester) fetch(u, method string, depth int) error {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return err
	}

	if err := h.checkRobots(parsedURL); err != nil {
		return err
	}

	if err := h.checkFilters(parsedURL); err != nil {
		return err
	}

	if err := h.checkDepth(depth); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(h.Context, method, parsedURL.String(), http.NoBody)
	if err != nil {
		return err
	}

	request := &Request{
		URL:       req.URL,
		Headers:   &req.Header,
		Host:      req.URL.Host,
		Method:    req.Method,
		Body:      req.Body,
		Depth:     depth,
		harvester: h,
	}

	h.handleRequestDo(request)

	res, err := h.Client.Do(req)
	if err != nil {
		return err
	}

	h.store.Visit(req.URL.String())

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

	// Reset the body reader for later use in `ResponseDo`.
	_, err = body.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	response := &Response{
		StatusCode: res.StatusCode,
		Headers:    &res.Header,
		Request:    request,
		Body:       body,
	}

	h.handleResponseDo(response)

	h.handleHtmlDo(response)

	return nil
}

func (h *Harvester) handleRequestDo(req *Request) {
	for _, m := range h.requestMiddlewares {
		m(req)
	}
}

func (h *Harvester) handleResponseDo(res *Response) {
	for _, m := range h.responseMiddlewares {
		m(res)
	}
}

func (h *Harvester) handleHtmlDo(res *Response) {
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Printf("error parsing response body: %v", err)
		return
	}

	for _, m := range h.htmlMiddlewares {
		doc.Find(m.Selector).Each(func(i int, s *goquery.Selection) {
			for _, n := range s.Nodes {
				el := &HtmlElement{
					attributes: n.Attr,
					Text:       s.Text(),
					Request:    res.Request,
					Response:   res,
					Selection:  s,
				}

				m.Function(el)
			}
		})
	}
}

func (h *Harvester) checkRobots(parsedURL *url.URL) error {
	if h.ignoreRobots {
		return nil
	}

	h.mu.Lock()
	robot, ok := h.robotsMap[parsedURL.Host]
	h.mu.Unlock()

	if !ok {
		robotURL := parsedURL.Scheme + "://" + parsedURL.Host + "/robots.txt"
		res, err := h.Client.Get(robotURL) //nolint: noctx // we don't need a context here
		if err != nil {
			return err
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				log.Printf("error closing response body: %v for request of: %v", err, robotURL)
			}
		}()

		robot, err = robotstxt.FromResponse(res)
		if err != nil {
			return err
		}

		h.mu.Lock()
		h.robotsMap[parsedURL.Host] = robot
		h.mu.Unlock()
	}

	if !robot.TestAgent(parsedURL.Path, "Grawlr") {
		return ErrRobotsDisallowed(parsedURL.String())
	}

	return nil
}

func (h *Harvester) checkFilters(parsedURL *url.URL) error {
	u := parsedURL.String()

	if !h.AllowRevisit && h.store.Visited(u) {
		return ErrVisitedURL(u)
	}

	if !h.isURLAllowed(u) {
		return ErrForbiddenURL(u)
	}

	return nil
}

func (h *Harvester) checkDepth(depth int) error {
	if h.DepthLimit != 0 && depth >= h.DepthLimit {
		return ErrDepthLimitExceeded(depth, h.DepthLimit)
	}

	return nil
}

// isURLAllowed checks if the given URL is allowed to be fetched.
func (h *Harvester) isURLAllowed(u string) bool {
	for _, disallowed := range h.DisallowedURLs {
		if strings.HasPrefix(u, disallowed) {
			return false
		}
	}

	if len(h.AllowedURLs) == 0 {
		return true
	}

	for _, allowed := range h.AllowedURLs {
		if strings.HasPrefix(u, allowed) {
			return true
		}
	}

	return false
}
