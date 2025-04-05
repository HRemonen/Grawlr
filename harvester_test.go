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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var helloBytes = []byte("Hello, client\n")

func newUnstartedTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(helloBytes)
	})

	mux.HandleFunc("/heavyweight", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(time.Second * 2): // Simulate work
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello"))
		case <-r.Context().Done(): // Handle request cancellation
			http.Error(w, "Request canceled", http.StatusRequestTimeout)
		}
	})

	mux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	})

	mux.Handle("/404", http.NotFoundHandler())

	mux.Handle("/allowed", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Allowed"))
	}))

	mux.Handle("/disallowed", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Disallowed"))
	}))

	mux.Handle("/robots.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User-agent: *\nDisallow: /disallowed"))
	}))

	mux.Handle("/user_agent", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(r.Header.Get("User-Agent")))
	}))

	mux.HandleFunc("/faq", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>FAQ</title></head>
			<body>
				<h1>Frequently Asked Questions</h1>
				<p>Welcome to the FAQ page. Here are some useful links:</p>
				<ul>
					<li><a href="/">Home</a></li>
					<li><a href="/about">About Us</a></li>
					<li><a href="/contact">Contact</a></li>
					<li><a href="/faq#section2">FAQ Section 2</a></li>
					<li><a href="https://external.com/resource">External Resource</a></li>
				</ul>
			</body>
			</html>
		`)
	})

	mux.HandleFunc("/relative_links", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Relative Links</title></head>
			<body>
				<h1>Relative Links Page</h1>
				<p>This page contains various types of relative links and whitespace:</p>
				<ul>
					<li><a href="/page1">Page 1</a></li>
					<li><a href="../page2">Page 2</a></li>
					<li><a href="./page3">Page 3</a></li>
					<li><a href="/path/to/page4">Nested Page 4</a></li>
					<li><a href="/path/to/page5#section1">Nested Page 5 with Anchor</a></li>
				</ul>
			</body>
			</html>
		`)
	})

	mux.HandleFunc("/complex_whitespace", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `
			<!DOCTYPE html>
			<html>
			<head><title>Complex Whitespace</title></head>
			<body>
				<h1>Complex Whitespace Page</h1>
				<p>This page contains tabs, newlines, and various spacing:</p>
				<p>Text with		tabs		and
				newlines</p>
				<p>
					Another paragraph with
					mixed spacing	and newlines.
				</p>
				<a href="/spaced_link">Spaced Link</a>
			</body>
			</html>
		`)
	})

	return httptest.NewUnstartedServer(mux)
}

func newTestServer() *httptest.Server {
	server := newUnstartedTestServer()
	server.Start()

	return server
}

func newTestHarvester(options ...Options) *Harvester {
	client := &http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return NewHarvester(
		append(options, WithClient(client))...,
	)
}

func TestHarvester_Visit(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	RequestDoCalled := false
	ResponseDoCalled := false

	f := newTestHarvester()

	f.RequestDo(func(req *Request) {
		RequestDoCalled = true
		req.Headers.Set("User-Agent", "Test User Agent")
	})

	f.ResponseDo(func(res *Response) {
		ResponseDoCalled = true

		assert.Equal(t, server.URL+"/", res.Request.URL.String())

		assert.Equal(t, "Test User Agent", res.Request.Headers.Get("User-Agent"))

		assert.Equal(t, http.StatusOK, res.StatusCode)

		bodyBytes, err := io.ReadAll(res.Body)

		assert.NoError(t, err)
		assert.Equal(t, helloBytes, bodyBytes)
	})

	f.Visit(server.URL + "/")

	if !RequestDoCalled {
		t.Error("RequestDo middleware was not called")
	}

	if !ResponseDoCalled {
		t.Error("ResponseDo middleware was not called")
	}
}

func TestHarvester_VisitRedirect(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	ResponseDoCalled := false

	f := newTestHarvester()

	f.ResponseDo(func(res *Response) {
		ResponseDoCalled = true

		assert.Equal(t, server.URL+"/redirect", res.Request.URL.String())
		assert.Equal(t, http.StatusSeeOther, res.StatusCode)
	})

	f.Visit(server.URL + "/redirect")

	if !ResponseDoCalled {
		t.Error("ResponseDo middleware was not called")
	}
}

func TestHarvester_VisitWithAllowedURLs(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	allowed := []string{
		server.URL + "/allowed",
		server.URL + "/faq",
	}

	f := newTestHarvester(WithAllowedURLs(allowed), WithIgnoreRobots(true))

	url := server.URL + "/"
	err := f.Visit(url)
	assert.EqualError(t, err, fmt.Sprintf("URL %s is forbidden", url))

	url = server.URL + "/disallowed"
	err = f.Visit(url)
	assert.EqualError(t, err, fmt.Sprintf("URL %s is forbidden", url))
}

func TestHarvester_VisitWithDisallowedURLs(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	disallowed := []string{
		server.URL + "/allowed",
		server.URL + "/faq",
	}

	ResponseDoCalled := false

	f := newTestHarvester(WithDisallowedURLs(disallowed))

	f.ResponseDo(func(res *Response) {
		ResponseDoCalled = true

		assert.Equal(t, http.StatusOK, res.StatusCode)
	})

	url := server.URL + "/allowed"
	err := f.Visit(url)
	assert.EqualError(t, err, fmt.Sprintf("URL %s is forbidden", url))

	url = server.URL + "/faq"
	err = f.Visit(url)
	assert.EqualError(t, err, fmt.Sprintf("URL %s is forbidden", url))

	url = server.URL + "/"
	err = f.Visit(url)
	assert.NoError(t, err)

	if !ResponseDoCalled {
		t.Error("ResponseDo middleware was not called")
	}
}

func TestHarvester_VisitWithContext(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	time.AfterFunc(time.Second, cancel)

	f := newTestHarvester(WithContext(ctx))

	f.ResponseDo(func(res *Response) {
		t.Error("ResponseDo middleware should not be called")
	})

	err := f.Visit(server.URL + "/heavyweight")

	// Since http.Client.Do returns an url.Error we need to cast the error to *url.Error
	// to check if the error is of type context.Canceled
	err, ok := err.(*url.Error)
	if !ok {
		t.Error("error is not of type *url.Error - which is expected from http.Client.Do")
	}
	assert.ErrorIs(t, err, context.Canceled)
}

func TestHarvester_MaximumDepth(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	h1 := newTestHarvester(WithDepthLimit(2), WithAllowRevisit(true))

	reqCount := 0
	h1.ResponseDo(func(resp *Response) {
		reqCount++
		if reqCount >= 10 {
			return
		}
		h1.Visit(server.URL + "/") // h.Visit does not increment the depth
	})

	h1.Visit(server.URL + "/")
	if reqCount < 10 {
		t.Errorf("Invalid number of request: %d (expected 10) without depth limit", reqCount)
	}

	h2 := newTestHarvester(WithDepthLimit(2), WithAllowRevisit(true))

	reqCount = 0
	h2.ResponseDo(func(resp *Response) {
		reqCount++
		resp.Request.Visit(server.URL + "/") // resp.Request.Visit increments the depth
	})

	h2.Visit(server.URL + "/")
	assert.Equal(t, 2, reqCount)
}
