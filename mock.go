package main

import "net/http"

// Mock contains mock info
type Mock struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Proxy    string   `json:"proxy"`
}

// MockCollection for work with mocks
type MockCollection struct {
	mocks []Mock
}

// Append mock to mock collection
func (c *MockCollection) Append(m []Mock) *MockCollection {
	c.mocks = append(c.mocks, m...)

	return c
}

// Lookup mock that looks like http request
func (c *MockCollection) Lookup(r *http.Request) *Mock {
	for _, mock := range collection.mocks {
		if mock.Request.LooksLike(r) {
			return &mock
		}
	}
	return nil
}
