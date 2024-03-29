package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Mock contains mock info
type Mock struct {
	Name     string   `json:"name"`
	Request  Request  `json:"request"`
	Response Response `json:"response"`
	Proxy    string   `json:"proxy"`
}

// MockCollection for work with mocks
type MockCollection struct {
	mutex sync.Mutex
	mocks []*Mock
}

// Lookup mock that looks like http request
func (c *MockCollection) Lookup(r *http.Request) *Mock {
	for _, mock := range collection.mocks {
		if mock.Request.LooksLike(r) {
			return mock
		}
	}
	return nil
}

func (c *MockCollection) Rebuild(files []string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.mocks = []*Mock{}
	for _, f := range files {
		temp, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("Unable to read %s file", f)
		}

		var mocks []*Mock

		if err := json.Unmarshal(temp, &mocks); err != nil {
			return fmt.Errorf("Unable to parse %s file", f)
		}

		for _, m := range mocks {
			if m.Name == "" {
				m.Name = fmt.Sprintf("%s: %s %s", f, m.Request.Method, m.Request.URL)
			}
		}

		for _, m := range mocks {
			if m.Response.File != nil {
				mockDir := filepath.Dir(f)
				absPath, err := filepath.Abs(fmt.Sprintf("%s/%s", mockDir, *m.Response.File))
				if err != nil {
					return err
				}

				if _, err := os.Stat(absPath); os.IsNotExist(err) {
					return fmt.Errorf("[%s] Data file %s not exists", m.Name, absPath)
				}

				m.Response.File = &absPath
			}
		}

		c.mocks = append(c.mocks, mocks...)
	}

	sort.Slice(c.mocks, func(i, j int) bool {
		return c.mocks[i].Request.Priority > c.mocks[j].Request.Priority
	})

	log.Println("Mocks found:", len(c.mocks))

	return nil
}
