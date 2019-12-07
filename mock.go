package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
)

// Mock contains mock info
type Mock struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Proxy    string   `json:"proxy"`
}

// MockCollection for work with mocks
type MockCollection struct {
	mutex sync.Mutex
	mocks []Mock
}

// Append mock to mock collection
func (c *MockCollection) Append(m []Mock) *MockCollection {
	c.mutex.Lock()
	c.mocks = append(c.mocks, m...)
	c.mutex.Unlock()

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

func (c *MockCollection) Rebuild(files []string) {
	c.mutex.Lock()

	c.mocks = []Mock{}
	for _, f := range files {
		temp, _ := ioutil.ReadFile(f)
		var mocks []Mock

		if err := json.Unmarshal(temp, &mocks); err != nil {
			log.Printf("Unable to parse %s file", f)
		}

		c.mocks = append(c.mocks, mocks...)
	}

	log.Println("Mocks found:", len(c.mocks))

	c.mutex.Unlock()
}
