package main

import (
	"bytes"
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// ProcessMock writes from mock to ResponseWriter
func ProcessMock(w http.ResponseWriter, r *http.Request, mock *Mock) error {
	if mock.Proxy != "" {
		return proxyRequest(w, r, mock)
	}

	if nil != mock.Response.Headers {
		for key, value := range *mock.Response.Headers {
			w.Header().Set(key, value)
		}
	}

	if mock.Response.Body != nil {
		if mock.Response.Code != nil {
			w.WriteHeader(*mock.Response.Code)
		}

		_, err := w.Write([]byte(*mock.Response.Body))
		if err != nil {
			return err
		}

		return nil
	}

	if mock.Response.File != nil {
		temp, err := os.ReadFile(*mock.Response.File)
		if err != nil {
			return err
		}

		if w.Header().Get("Content-Type") == "" {
			switch filepath.Ext(*mock.Response.File) {
			case ".json":
				w.Header().Set("Content-Type", "application/json")
			case ".html":
				w.Header().Set("Content-Type", "text/html")
			case ".xml":
				w.Header().Set("Content-Type", "application/xml")
			}
		}

		if mock.Response.Code != nil {
			w.WriteHeader(*mock.Response.Code)
		}

		_, err = w.Write(temp)
		if err != nil {
			return err
		}

		return nil
	}

	if mock.Response.JSON != nil {
		if w.Header().Get("Content-Type") == "" {
			w.Header().Set("Content-Type", "application/json")
		}

		if mock.Response.Code != nil {
			w.WriteHeader(*mock.Response.Code)
		}

		_, err := w.Write(*mock.Response.JSON)
		if err != nil {
			return err
		}

		// if mock.Response.Code != nil {
		// 	w.WriteHeader(*mock.Response.Code)
		// }
		return nil
	}

	return nil
}

func proxyRequest(w http.ResponseWriter, r *http.Request, mock *Mock) error {
	body := getBodyCopy(r)
	pr, _ := http.NewRequest(r.Method, r.URL.String(), io.NopCloser(bytes.NewBuffer(body)))
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
	_, err = w.Write(pb)
	if err != nil {
		return err
	}

	return nil
}
