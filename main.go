package main

import (
  "context"
  "crypto/tls"
  "fmt"
  "log"
  "net/http"
  "sync"
  "time"
  "os"
  "path/filepath"

  "github.com/caarlos0/env"
  "github.com/joho/godotenv"
  "golang.org/x/crypto/acme/autocert"
)

type config struct {
  Domains []string `env:"DOMAINS" envDefault:"localhost"`
  Port    string   `env:"PORT" envDefault:"3000"`
  UseSSL  bool     `env:"USE_SSL" envDefault:"false"`
  SSLPort string   `env:"SSL_PORT" envDefault:"443"`
  CertDir string   `env:"CERT_DIR" envDefault:"."`
  DataDir string   `env:"DATA_DIR" envDefault:"data/"`
}

func main() {
  if err := godotenv.Load(); err != nil {
    log.Println("File .env not found, reading configuration from ENV")
  }

  var cfg config
  if err := env.Parse(&cfg); err != nil {
    log.Fatalln("Failed to parse ENV")
  }
  if absDir, err := filepath.Abs(cfg.DataDir); err == nil {
    cfg.DataDir = absDir
  } else {
    cwd, err := os.Getwd()
    if err != nil {
      log.Fatal(err)
    }
    cfg.DataDir = cwd
  }

  log.Println(cfg)

  handleFunc := CreateFileHandler(cfg)

  mux := &http.ServeMux{}
  mux.HandleFunc("/", handleFunc)

  var wg sync.WaitGroup
  wg.Add(2)

  go func() {
    defer wg.Done()
    server := &http.Server{
      ReadTimeout:  5 * time.Second,
      WriteTimeout: 5 * time.Second,
      IdleTimeout:  120 * time.Second,
      Handler:      mux,
      Addr:         ":" + cfg.Port,
    }
    log.Fatal(server.ListenAndServe())
  }()

  go func() {
    defer wg.Done()
    if !cfg.UseSSL {
      log.Println("Skipping ssl")
      return
    }

    hostPolicy := func(ctx context.Context, host string) error {
      for _, domain := range cfg.Domains {
        if domain == host {
          return nil
        }
      }
      return fmt.Errorf("acme/autocert: host %s is not allowed", host)
    }

    certManager := &autocert.Manager{
      Prompt:     autocert.AcceptTOS,
      HostPolicy: hostPolicy,
      Cache:      autocert.DirCache("."),
    }

    server := &http.Server{
      ReadTimeout:  5 * time.Second,
      WriteTimeout: 5 * time.Second,
      IdleTimeout:  120 * time.Second,
      Handler:      mux,
      Addr:         ":" + cfg.SSLPort,
      TLSConfig:    &tls.Config{GetCertificate: certManager.GetCertificate},
    }
    log.Fatal(server.ListenAndServeTLS("", ""))
  }()

  wg.Wait()
}
