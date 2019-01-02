package main

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProcessMock(t *testing.T) {
	req := httptest.NewRequest(
		"POST",
		"/api",
		strings.NewReader("{\"name\":\"John\"}"),
	)

	rr := httptest.NewRecorder()
	code := 201
	json := json.RawMessage(`{"status":"ok"}`)
	response := Response{
		Code: &code,
		JSON: &json,
	}
	_ = ProcessMock(rr, req, &Mock{Response: response})

	if status := rr.Code; status != 201 {
		t.Errorf(
			"handler returned wrong status code: got %v want %v",
			status, 201,
		)
	}
	if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
		t.Errorf("content type header does not match: got %v want %v",
			ctype, "application/json")
	}
	if !bytes.Equal(rr.Body.Bytes(), json) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.Bytes(), json)
	}
}
