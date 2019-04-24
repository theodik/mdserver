package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func sanitizePath(url string) string {
	path := url
	path = path[1:]
	if path == "" {
		path = "index.html"
	}
	if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	}
	if filepath.Ext(path) == "" {
		path = path + ".html"
	}
	return path
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, `<!doctype html><meta charset="utf-8"><h1>Internal server error</h1><hr>`)
	io.WriteString(w, fmt.Sprintf("Error: %s", err))
}

// CreateFileHandler creates handler for handling files duuh
func CreateFileHandler(cfg config) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, `<!doctype html><meta charset="utf-8"><h1>Page not found</h1>`)
			return
		}

		path := sanitizePath(r.URL.Path)
		filename := filepath.Join(cfg.DataDir, path)
		if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, fmt.Sprintf(`<!doctype html><meta charset="utf-8"><h1>File '%s' not found</h1>`, filename))
			return
		} else if err != nil {
			writeError(w, err)
			return
		}

		data, err := ioutil.ReadFile(filename)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, `<!doctype html><meta charset="utf-8"><h1>Internal server error</h1><hr>`)
			io.WriteString(w, fmt.Sprintf("Error: %s", err))
		}

		io.WriteString(w, `<!doctype html><meta charset="utf-8">`)
		w.Write(data)
	}
}
