package main

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// ProcessMock writes from mock to ResponseWriter
func ProcessMock(w http.ResponseWriter, r *http.Request, mock *Mock) error {
	body := GetBodyCopy(r)

	if mock.Proxy != "" {
		pr, _ := http.NewRequest(r.Method, r.URL.String(), ioutil.NopCloser(bytes.NewBuffer(body)))
		pr.Header = r.Header

		pr.URL.Host = mock.Proxy
		pr.URL.Scheme = r.URL.Scheme
		if strings.Contains(mock.Proxy, "://") {
			purl, err := url.Parse(mock.Proxy)
			if err != nil {
				return err
			}

			pr.URL.Host = purl.Host
			pr.URL.Scheme = purl.Scheme
		}

		if pr.URL.Scheme == "" {
			pr.URL.Scheme = "http"
		}

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}
		resp, err := client.Do(pr)
		if err != nil {
			return err
		}

		w.WriteHeader(resp.StatusCode)
		for k := range resp.Header {
			w.Header().Set(k, resp.Header.Get(k))
		}

		pb, _ := ioutil.ReadAll(resp.Body)
		w.Write(pb)

		return nil
	}

	if nil != mock.Response.Headers {
		for key, value := range *mock.Response.Headers {
			w.Header().Set(key, value)
		}
	}

	if mock.Response.JSON != nil && mock.Response.Body == nil {
		w.Header().Set("Content-Type", "application/json")
	}

	if mock.Response.Code != nil {
		w.WriteHeader(*mock.Response.Code)
	}

	if mock.Response.Body != nil {
		w.Write([]byte(*mock.Response.Body))

		return nil
	}

	if mock.Response.JSON != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write(*mock.Response.JSON)

		return nil
	}

	return nil
}
