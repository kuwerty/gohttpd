package main;

import (
  "log"
  "net/http"
  "flag"
)

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

    files.ServeHTTP(&dw, r);

    h := dw.Header()

    log.Printf(" %d %s %s %s %s\n", dw.Code, r.Method, r.RequestURI, h.Get("Content-Type"), h.Get("Content-Length"))

  }

  return http.HandlerFunc(foo)
}


func main() {
  addr := flag.String("addr", ":8000", "listen address")
  root := flag.String("root", ".", "www root directory")
  flag.Parse()

  log.Printf("Ready on %s with root %s\n", *addr, *root)

  log.Fatal(http.ListenAndServe(*addr, mainHandler(*root)))
}

