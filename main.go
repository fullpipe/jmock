package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
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
		Version: "0.2.0",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args[0])
			files, err := filepath.Glob(args[0])
			if err != nil {
				log.Fatal(err)
			}

			err = collection.Rebuild(files)
			if err != nil {
				log.Fatal(err)
			}

			if Watch {
				go watchAndRebuildCollection(files)
			}

			http.HandleFunc("/", errorHandler(httpHandler))
			server := &http.Server{
				Addr:              fmt.Sprintf(":%d", Port),
				ReadHeaderTimeout: 3 * time.Second,
			}

			err = server.ListenAndServe()
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

	sem := semaphore.NewWeighted(1)
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
					defer sem.Release(1)

					log.Println("Changes detected. Updating mocks...")
					time.Sleep(100 * time.Millisecond)
					errRebuild := collection.Rebuild(files)
					if errRebuild != nil {
						log.Println("Error while rebuilding collection:\n", errRebuild)
					}
				}()
			case errWatch, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("Watch error:", errWatch)
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
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	if AllowOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", AllowOrigin)
	}

	if mock == nil {
		log.Printf("Mock not found for request: %s %s\n", r.Method, r.URL.Path)
		w.WriteHeader(501)
		return nil
	}

	return errors.Wrap(ProcessMock(w, r, mock), mock.Name)
}

func getBodyCopy(req *http.Request) []byte {
	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

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
