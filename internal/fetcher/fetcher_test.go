package fetcher

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

func newTestFetcher() *HTTPFetcher {
	return NewHTTPFetcher(&http.Client{
		Timeout: time.Second * 10,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	})
}

func TestHTTPFetcher_FetchHomePage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, client\n", string(body))
}

func TestHTTPFetcher_FetchRedirect(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/redirect")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusSeeOther, resp.StatusCode)
	assert.Equal(t, "/", resp.Headers.Get("Location"))
}

func TestHTTPFetcher_FetchErrorPage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/error")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "Internal server error")
}

func TestHTTPFetcher_FetchNotFoundPage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/404")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHTTPFetcher_FetchWithRobotsAllowed(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/allowed")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "Allowed", string(body))
}

func TestHTTPFetcher_FetchWithRobotsDisallowed(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	_, err := f.Fetch(server.URL + "/disallowed")
	assert.ErrorIs(t, err, ErrRobotsDisallowed)
}

func TestHTTPFetcher_FetchRobotsTxt(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/robots.txt")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(body), "User-agent: *\nDisallow: /disallowed")
}

func TestHTTPFetcher_FetchFAQPage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/faq")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	content := string(body)

	assert.Contains(t, content, `<h1>Frequently Asked Questions</h1>`)
	assert.Contains(t, content, `<a href="/">Home</a>`)
	assert.Contains(t, content, `<a href="/about">About Us</a>`)
	assert.Contains(t, content, `<a href="https://external.com/resource">External Resource</a>`)
}

func TestHTTPFetcher_FetchRelativeLinksPage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/relative_links")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	content := string(body)

	assert.Contains(t, content, `<a href="/page1">Page 1</a>`)
	assert.Contains(t, content, `<a href="../page2">Page 2</a>`)
	assert.Contains(t, content, `<a href="./page3">Page 3</a>`)
}

func TestHTTPFetcher_FetchComplexWhitespacePage(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	f := newTestFetcher()
	resp, err := f.Fetch(server.URL + "/complex_whitespace")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	content := string(body)

	assert.Contains(t, content, `<h1>Complex Whitespace Page</h1>`)
	assert.Contains(t, content, `<a href="/spaced_link">Spaced Link</a>`)
}
