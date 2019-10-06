package main

import "encoding/json"

// Response describes server response for mock
type Response struct {
	Code    *int               `json:"code,omitempty"`
	Body    *string            `json:"body,omitempty"`
	JSON    *json.RawMessage   `json:"json,omitempty"`
	Headers *map[string]string `json:"headers,omitempty"`
}
