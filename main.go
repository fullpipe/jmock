package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var Port int
var Watch bool
var AllowOrigin string

var collection MockCollection

func main() {
	var rootCmd = &cobra.Command{
		Use:     "jmock <path to mocks>",
		Short:   "Simple and easy to use json/post API mock server",
		Long:    `Simple and easy to use json/post API mock server`,
		Version: "0.1.0",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			files, err := filepath.Glob(args[0])
			if err != nil {
				log.Fatal(err)
			}

			collection.Rebuild(files)

			if Watch {
				go watchAndRebuildCollection(files)
			}

			http.HandleFunc("/", errorHandler(httpHandler))
			log.Println("Listening on port", fmt.Sprintf(":%d", Port))
			err = http.ListenAndServe(fmt.Sprintf(":%d", Port), nil)
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	rootCmd.Flags().IntVarP(&Port, "port", "p", 9090, "Specify port to listen")
	rootCmd.Flags().BoolVarP(&Watch, "watch", "w", false, "Watch for file changes")
	rootCmd.Flags().StringVarP(&AllowOrigin, "allow-origin", "o", "*", "Set up Access-Control-Allow-Origin header")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func watchAndRebuildCollection(files []string) {
	log.Printf("Watching for %d files\n", len(files))
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	var sem = semaphore.NewWeighted(1)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Fatal("Watcher`s event failed")
				}

				if event.Op.String() != "WRITE" {
					continue
				}

				if !sem.TryAcquire(1) {
					continue
				}

				go func() {
					log.Println("Changes detected. Updating mocks...")
					time.Sleep(100 * time.Millisecond)
					collection.Rebuild(files)
					sem.Release(1)
				}()
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watch error:", err)
			}
		}
	}()

	for _, f := range files {
		err = watcher.Add(f)
		if err != nil {
			log.Fatal(err)
		}
	}

	<-done
}

func httpHandler(w http.ResponseWriter, r *http.Request) error {
	body := getBodyCopy(r)

	mock := collection.Lookup(r)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	if AllowOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", AllowOrigin)
	}

	if mock == nil {
		log.Println("Mock not found for request")
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
			log.Printf("Error %q: %v", r.RequestURI, err)
		}
	}
}
