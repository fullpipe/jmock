package main

import "encoding/json"

type Response struct {
	Code *int             `json:"code,omitempty"`
	Body *string          `json:"body,omitempty"`
	JSON *json.RawMessage `json:"json,omitempty"`
}
