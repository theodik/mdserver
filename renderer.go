package main

import (
  "fmt"
  "log"
  "io"
  "io/ioutil"
  "net/http"
  "os"
  "path/filepath"
  "strings"

  "github.com/russross/blackfriday"
  "github.com/cyphar/filepath-securejoin"
)

// sanitizePath makes full path to file
// "" => /index.{ext}
// "/" => /index.{ext}
// "/neco" => /neco.{ext}
func sanitizePath(url string, ext string) string {
  path := url
  path = path[1:]
  if path == "" {
    path = "index." + ext
  }
  if strings.HasSuffix(path, "/") {
    path = path + "index." + ext
  }
  if filepath.Ext(path) == "" {
    path = path + "." + ext
  }
  return path
}

func fileExists(path string) bool {
  _, err := os.Stat(path)
  if err == nil {
    return true
  }
  if os.IsNotExist(err) {
    return false
  }
  return false
}

func findFile(basepath string, path string, exts []string) (string, string, bool) {
  if ext := filepath.Ext(path); ext != "" {
    filename, err := securejoin.SecureJoin(basepath, path)
    if err != nil {
      return filename, ext, false
    }
    if fileExists(filename) {
      return filename, ext, true
    }
    return filename, ext, false
  }

  for _, ext := range exts {
    sanitizedPath := sanitizePath(path, ext)
    filename, err := securejoin.SecureJoin(basepath, sanitizedPath)
    if err != nil {
      return filename, ext, false
    }

    if fileExists(filename) {
      return filename, ext, true
    } 
  }
  return path, "", false
}

func writeError(w http.ResponseWriter, err error) {
  w.WriteHeader(http.StatusInternalServerError)
  io.WriteString(w, `<!doctype html><meta charset="utf-8"><h1>Internal server error</h1><hr>`)
  io.WriteString(w, fmt.Sprintf("Error: %s", err))
}

// CreateFileHandler creates handler for handling files
func CreateFileHandler(cfg config) func(http.ResponseWriter, *http.Request) {
  return func(w http.ResponseWriter, r *http.Request) {
    if r.Method != "GET" {
      w.WriteHeader(http.StatusNotFound)
      io.WriteString(w, `<!doctype html><meta charset="utf-8"><h1>Page not found</h1>`)
      log.Println(r.Method, r.URL.Path, "=> Invalid http method")
      return
    }

    filename, ext, found := findFile(cfg.DataDir, r.URL.Path, []string{"html", "md"})
    if !found {
      w.WriteHeader(http.StatusNotFound)
      io.WriteString(w, fmt.Sprintf(`<!doctype html><meta charset="utf-8"><h1>404 Not found</h1>`))
      log.Println(r.Method, r.URL.Path, "=> 404 Not found")
      return
    }

    data, err := ioutil.ReadFile(filename)
    if err != nil {
      w.WriteHeader(http.StatusInternalServerError)
      io.WriteString(w, `<!doctype html><meta charset="utf-8"><h1>Internal server error</h1><hr>`)
      io.WriteString(w, fmt.Sprintf("Error: %s", err))
      log.Println(r.Method, r.URL.Path, "=>", err)
      return
    }

    io.WriteString(w, `<!doctype html><meta charset="utf-8">`)
    if ext == "md" {
      html := blackfriday.Run(data)
      w.Write(html)
    } else {
      w.Write(data)
    }

    log.Println(r.Method, r.URL.Path, "=>", filename)
  }
}
