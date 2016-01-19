package main;

import (
  "log"
  "net/http"
  "flag"
)

var cache = flag.Bool("cache", true, "enable cached response headers")
var debug = flag.Bool("debug", false, "enable debugging")
var secure = flag.Bool("secure", false, "server over TLS (requires server.key, server.pem)")

type DebugWriter struct {
  http.ResponseWriter
  Code int
}

func (w *DebugWriter) WriteHeader(code int) {
  w.Code = code
  w.ResponseWriter.WriteHeader(code)
}


func mainHandler(root string) http.Handler {
  files := http.FileServer(http.Dir(root))

  foo := func(w http.ResponseWriter, r *http.Request) {

    dw := DebugWriter{ResponseWriter:w}

    w.Header().Set("Cache-Control", "no-cache")

    w.Header().Set("Access-Control-Allow-Origin", "*")

    files.ServeHTTP(&dw, r);

    h := dw.Header()

    log.Printf("%d %s %s %s %s\n", dw.Code, r.Method, r.RequestURI, h.Get("Content-Type"), h.Get("Content-Length"))

    if *debug {
      for k,v := range(r.Header) {
        log.Printf("  %s: %s\n", k, v)
      }
    }
  }

  return http.HandlerFunc(foo)
}


func main() {
  addr := flag.String("addr", ":8000", "listen address")
  root := flag.String("root", ".", "www root directory")
  flag.Parse()

  log.Printf("Ready on %s with root %s\n", *addr, *root)

  if(*secure) {
    log.Printf("Serving TLS")
    log.Fatal(http.ListenAndServeTLS(*addr, "server.pem", "server.key", mainHandler(*root)))
  }

  log.Fatal(http.ListenAndServe(*addr, mainHandler(*root)))
}

