package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var collection MockCollection

func main() {
	var Port int
	var rootCmd = &cobra.Command{
		Use:     "jmock <path to mocks>",
		Short:   "Simple and easy to use json/post API mock server",
		Long:    `Simple and easy to use json/post API mock server`,
		Version: "0.0.1",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			files, err := filepath.Glob(args[0])
			if err != nil {
				log.Fatal(err)
			}

			buildCollection(files)

			http.HandleFunc("/", errorHandler(httpHandler))
			log.Println("Listening on port", fmt.Sprintf(":%d", Port))
			err = http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.Flags().IntVarP(&Port, "port", "p", 9090, "Specify port to listen")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func buildCollection(files []string) {
	for _, f := range files {
		temp, _ := ioutil.ReadFile(f)
		var mocks []Mock

		if err := json.Unmarshal(temp, &mocks); err != nil {
			log.Printf("Unable to parse %s file", f)
		}

		collection.Append(mocks)

		log.Println("Mocks found:", len(collection.mocks))
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) error {
	body := getBodyCopy(r)

	mock := collection.Lookup(r)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if mock == nil {
		log.Println("No mock found for request")
		w.WriteHeader(501)
		return nil
	}

	return ProcessMock(w, r, mock)
}

func getBodyCopy(req *http.Request) []byte {
	bodyBytes, _ := ioutil.ReadAll(req.Body)
	req.Body.Close()
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes
}

func errorHandler(f func(http.ResponseWriter, *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error handling %q: %v", r.RequestURI, err)
		}
	}
}
