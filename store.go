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

import "sync"

// Storer is an interface for a cache that storer
// Harvester's internal data.
type Storer interface {
	// Visited returns true if the URL has been visited.
	Visited(url string) bool
	// Visit marks the URL as visited.
	Visit(url string)
}

type InMemoryStore struct {
	visited map[string]bool
	lock    *sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		visited: make(map[string]bool),
		lock:    &sync.RWMutex{},
	}
}

func (s *InMemoryStore) Visited(url string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.visited[url]
}

func (s *InMemoryStore) Visit(url string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.visited[url] = true
}
