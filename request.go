package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/buger/jsonparser"
	"github.com/gobwas/glob"
)

// Request describes mock request
type Request struct {
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url,omitempty"`
	Post    map[string]string `json:"post,omitempty"`
	Query   map[string]string `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	JSON    *json.RawMessage  `json:"json,omitempty"`
}

// LooksLike checks if mock request looks like real request
func (mockRequest Request) LooksLike(req *http.Request) bool {
	if err := req.ParseForm(); err != nil {
		log.Println("ParseForm() Error: ", err)
	}

	var g glob.Glob

	if mockRequest.URL != "" {
		g = glob.MustCompile(mockRequest.URL)
		if !g.Match(req.URL.RequestURI()) {
			return false
		}
	}

	if mockRequest.Method != "" {
		g = glob.MustCompile(mockRequest.Method)
		if !g.Match(req.Method) {
			return false
		}
	}

	for key, value := range mockRequest.Query {
		g = glob.MustCompile(value)
		if !g.Match(req.URL.Query().Get(key)) {
			return false
		}
	}

	for key, value := range mockRequest.Post {
		g = glob.MustCompile(value)
		if !g.Match(req.PostFormValue(key)) {
			return false
		}
	}

	for key, value := range mockRequest.Headers {
		g = glob.MustCompile(value)
		if !g.Match(req.Header.Get(key)) {
			return false
		}
	}

	if mockRequest.JSON != nil {
		_, dataType, _, err := jsonparser.Get(*mockRequest.JSON)
		if err != nil {
			log.Panicln(err)
		}

		body := GetBodyCopy(req)

		if !compareJSON(*mockRequest.JSON, body, dataType) {
			return false
		}
	}

	return true
}

func compareJSON(mock []byte, real []byte, dataType jsonparser.ValueType) bool {
	jsonMatches := true

	switch dataType {
	case jsonparser.Object:
		jsonparser.ObjectEach(mock, func(key []byte, value []byte, dataType jsonparser.ValueType, offset int) error {
			rvalue, _, _, err := jsonparser.Get(real, string(key))
			if err != nil {
				jsonMatches = false
			}

			//log.Println("compareObj: ", value, rvalue, dataType, compareJson(value, rvalue, dataType))
			if !compareJSON(value, rvalue, dataType) {
				jsonMatches = false
			}
			return nil
		})
	case jsonparser.Array:
		jsonparser.ArrayEach(mock, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			rvalue, _, _, err := jsonparser.Get(real, fmt.Sprintf("%s%d%s", "[", offset-1, "]"))
			if err != nil {
				jsonMatches = false
			}

			if !compareJSON(value, rvalue, dataType) {
				jsonMatches = false
			}
		})
	default:
		g := glob.MustCompile(string(mock))
		jsonMatches = g.Match(string(real))
	}

	return jsonMatches
}
