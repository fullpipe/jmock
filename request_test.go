package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type testcase struct {
	real   *http.Request
	mock   Request
	result bool
}

var urlTests = []testcase{
	{
		httptest.NewRequest("POST", "/api", nil),
		Request{Method: "POST"},
		true,
	},
	{
		httptest.NewRequest("POST", "/api", nil),
		Request{},
		true,
	},
	{
		httptest.NewRequest("POST", "/api", nil),
		Request{Method: "GET"},
		false,
	},
	{
		httptest.NewRequest("POST", "/foo", nil),
		Request{URL: "/foo"},
		true,
	},
	{
		httptest.NewRequest("POST", "/foo", nil),
		Request{URL: "/bar"},
		false,
	},
	{
		httptest.NewRequest("POST", "/foo", nil),
		Request{URL: "/foo2"},
		false,
	},
	{
		httptest.NewRequest("POST", "/foo", nil),
		Request{URL: "/f*"},
		true,
	},
}

func TestLooksLikeUrlAndMethod(t *testing.T) {
	for idx, tc := range urlTests {
		if tc.result != tc.mock.LooksLike(tc.real) {
			if tc.result {
				t.Error(
					"#", idx, ": Input", tc.real,
					"should looks like mock", tc.mock,
				)
			} else {
				t.Error(
					"#", idx, ": Input", tc.real,
					"should NOT looks like mock", tc.mock,
				)
			}
		}
	}
}

type postTestCase struct {
	post     url.Values
	mockData map[string]string
	result   bool
}

var postTests = []postTestCase{
	{
		url.Values{"foo": []string{"bar"}},
		map[string]string{"foo": "bar"},
		true,
	},
	{
		url.Values{
			"foo":  []string{"bar"},
			"foo2": []string{"bar2"},
		},
		map[string]string{
			"foo2": "b*",
		},
		true,
	},
	{
		url.Values{
			"foo":  []string{"bar"},
			"foo2": []string{"bar2"},
		},
		map[string]string{
			"foo":  "not bar",
			"foo2": "b*",
		},
		false,
	},
}

func TestLooksLikePostParams(t *testing.T) {
	for idx, tc := range postTests {
		real := httptest.NewRequest("POST", "/api", strings.NewReader(tc.post.Encode()))
		real.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		mock := Request{Post: tc.mockData}

		if tc.result != mock.LooksLike(real) {
			if tc.result {
				t.Error(
					"#", idx, ": Input", real,
					"should looks like mock", mock,
				)
			} else {
				t.Error(
					"#", idx, ": Input", real,
					"should NOT looks like mock", mock,
				)
			}
		}
	}
}

var jsonTests = []struct {
	json     string
	mockData string
	result   bool
}{
	{
		"{}",
		"{}",
		true,
	},
	{
		`{"foo":"rab"}`,
		`{"foo":"bar"}`,
		false,
	},
	{
		`[{"foo":"bar"}]`,
		`[{"foo":"bar"}]`,
		true,
	},
	{
		`{"foo":11}`,
		`{"foo":"1*"}`,
		true,
	},
	{
		`{"foo":"bar"}`,
		`{"foo":"bar"}`,
		true,
	},
	{
		`{"foo":{"f": "f", "foo2":"bar"}}`,
		`{"foo":{"foo2":"bar"}}`,
		true,
	},
	{
		`{"foo":{"f": "f", "foo2":"bar"}}`,
		`{"foo":{"foo2":"b*"}}`,
		true,
	},
}

func TestLooksLikeJson(t *testing.T) {
	for idx, tc := range jsonTests {
		real := httptest.NewRequest("POST", "/api", strings.NewReader(tc.json))
		real.Header.Add("Content-Type", "application/json")

		h := json.RawMessage(tc.mockData)
		mock := Request{JSON: &h}

		if tc.result != mock.LooksLike(real) {
			if tc.result {
				t.Error(
					"#", idx, ": Input", real,
					"should looks like mock", mock,
				)
			} else {
				t.Error(
					"#", idx, ": Input", real,
					"should NOT looks like mock", mock,
				)
			}
		}
	}
}

var headerTests = []struct {
	headers     map[string]string
	mockHeaders map[string]string
	result      bool
}{
	{
		nil,
		nil,
		true,
	},
	{
		map[string]string{"foo": "bar"},
		nil,
		true,
	},
	{
		map[string]string{"foo": "bar"},
		map[string]string{"foo": "bar"},
		true,
	},
	{
		map[string]string{"foo": "bar"},
		map[string]string{"foo": "ba*"},
		true,
	},
	{
		map[string]string{"foo": "bar"},
		map[string]string{"foo": "notbar"},
		false,
	},
	{
		map[string]string{"foo": "asdsbar"},
		map[string]string{"foo": "a*bar"},
		true,
	},
	{
		map[string]string{"foo": "foo"},
		map[string]string{"foo": ""},
		false,
	},
}

func TestLooksLikeHeaders(t *testing.T) {
	for idx, tc := range headerTests {
		real := httptest.NewRequest("POST", "/api", nil)

		for key, header := range tc.headers {
			real.Header.Add(key, header)
		}

		mock := Request{Method: "POST", Headers: tc.mockHeaders}

		if tc.result != mock.LooksLike(real) {
			if tc.result {
				t.Error(
					"#", idx, ": Input", real,
					"should looks like mock", mock,
				)
			} else {
				t.Error(
					"#", idx, ": Input", real,
					"should NOT looks like mock", mock,
				)
			}
		}
	}
}
