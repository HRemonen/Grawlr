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
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newUnstartedTestServer() *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, client\n"))
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

func newTestFetcher(options ...Options) *Fetcher {
	client := &http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return NewFetcher(
		append(options, WithClient(client))...,
	)
}

func TestFetcher_Visit(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	onRequestCalled := false

	f := newTestFetcher()

	f.OnRequest(func(req *Request) {
		onRequestCalled = true
		req.Headers.Set("User-Agent", "Test User Agent")
	})

	res, err := f.Visit(server.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	if !onRequestCalled {
		t.Error("OnRequest middleware was not called")
	}

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, client\n", string(body))
}

func TestFetcher_VisitRedirect(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/redirect")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, res.StatusCode)
	assert.Equal(t, "/", res.Headers.Get("Location"))
}

func TestFetcher_VisitErrorPage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/error")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Internal server error")
}

func TestFetcher_VisitNotFoundPage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/404")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestFetcher_VisitWithRobotsAllowed(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/allowed")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Allowed", string(body))
}

func TestFetcher_VisitWithRobotsDisallowed(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	_, err := f.Visit(server.URL + "/disallowed")
	assert.ErrorIs(t, err, ErrRobotsDisallowed)
}

func TestFetcher_VisitRobotsTxt(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/robots.txt")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "User-agent: *\nDisallow: /disallowed")
}

func TestFetcher_VisitRelativeLinks(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/relative_links")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	content := string(body)

	assert.Contains(t, content, `<a href="/page1">Page 1</a>`)
	assert.Contains(t, content, `<a href="../page2">Page 2</a>`)
	assert.Contains(t, content, `<a href="./page3">Page 3</a>`)
}

func TestFetcher_VisitComplexWhitespace(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	res, err := f.Visit(server.URL + "/complex_whitespace")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	body, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	content := string(body)

	assert.Contains(t, content, `<h1>Complex Whitespace Page</h1>`)
	assert.Contains(t, content, `<a href="/spaced_link">Spaced Link</a>`)
}

func TestFetcher_VisitWithAllowedURLs(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	allowed := []string{
		server.URL + "/allowed",
		server.URL + "/faq",
	}

	f := newTestFetcher(WithAllowedURLs(allowed))

	res, err := f.Visit(server.URL + "/allowed")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res, err = f.Visit(server.URL + "/faq")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	_, err = f.Visit(server.URL + "/")
	assert.ErrorIs(t, err, ErrForbiddenURL)
}

func TestFetcher_VisitWithDisallowedURLs(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	disallowed := []string{
		server.URL + "/allowed",
		server.URL + "/faq",
	}

	f := newTestFetcher(WithDisallowedURLs(disallowed))

	_, err := f.Visit(server.URL + "/allowed")
	assert.ErrorIs(t, err, ErrForbiddenURL)

	_, err = f.Visit(server.URL + "/faq")
	assert.ErrorIs(t, err, ErrForbiddenURL)

	res, err := f.Visit(server.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}
