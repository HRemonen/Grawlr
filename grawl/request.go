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
	"net/url"
	"strings"
)

type Request struct {
	URL     *url.URL
	BaseURL *url.URL
	Headers *http.Header
	Host    string
	Method  string
	Body    io.Reader
}

// GetAbsoluteURL returns the absolute URL for a link found on the page.
func (r *Request) GetAbsoluteURL(link string) string {
	if strings.HasPrefix(link, "#") {
		return ""
	}

	base, err := url.Parse(r.URL.String())
	if err != nil {
		fmt.Printf("Error parsing base URL: %s", err)
		return ""
	}

	href, err := url.Parse(link)
	if err != nil {
		fmt.Printf("Error parsing href: %s", err)
		return ""
	}

	absoluteURL := base.ResolveReference(href)
	return absoluteURL.String()
}
