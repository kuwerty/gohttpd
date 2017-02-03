package main;

import (
  "log"
  "net/http"
  "flag"
  "mime"
  "github.com/go-ini/ini"
)

var cache  = flag.Bool("cache", true, "enable cached response headers")
var debug  = flag.Bool("debug", false, "enable debugging")
var secure = flag.Bool("secure", false, "server over TLS (requires server.key, server.pem)")
var addr   = flag.String("addr", ":8000", "listen address")
var root   = flag.String("root", ".", "www root directory")

type DebugWriter struct {
  http.ResponseWriter
  Code int
}

func (w *DebugWriter) WriteHeader(code int) {
  w.Code = code
  w.ResponseWriter.WriteHeader(code)
}


func wrapHandler(name string, handler http.Handler) http.Handler {

  foo := func(w http.ResponseWriter, r *http.Request) {

    dw := DebugWriter{ResponseWriter:w}

    if(!*cache) {
      w.Header().Set("Cache-Control", "no-cache")
    } else {
      w.Header().Set("Cache-Control", "max-age=120")
    }

    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "*") //POST, PUT, GET, OPTIONS, HEAD")

    handler.ServeHTTP(&dw, r);

    h := dw.Header()

    log.Printf("%d %s %s (%s) %s %s\n", dw.Code, r.Method, r.RequestURI, name, h.Get("Content-Type"), h.Get("Content-Length"))

    if *debug {
      for k,v := range(r.Header) {
        log.Printf("  %s: %s\n", k, v)
      }
    }
  }

  return http.HandlerFunc(foo)
}



func main() {
  flag.Parse()

  certFile := "server.pem"
  keyFile := "server.key"

  mux := http.NewServeMux()

  var cfg, err = ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, "gohttpd.conf")

  if cfg != nil && err == nil {
    log.Printf("found a config file\n")

    var defaults = cfg.Section("")

    if defaults.HasKey("cert-file") {
      certFile = defaults.Key("cert-file").String()
    }

    if defaults.HasKey("key-file") {
      keyFile = defaults.Key("key-file").String()
    }

    var paths = cfg.Section("paths")

    for _,key := range(paths.KeyStrings()) {

      val := paths.Key(key).String()

      log.Printf("mapping %s to %s", key, val)

      handler := http.FileServer(http.Dir(val))
      handler = http.StripPrefix(key, handler)
      handler = wrapHandler(val, handler)

      mux.Handle(key, handler)
    }

  } else {

    mux.Handle("/", wrapHandler("/", http.FileServer(http.Dir(*root))))

  }


  mime.AddExtensionType(".aex", "application/json")

  log.Printf("Ready on %s with root %s\n", *addr, *root)

  if(*secure) {
    log.Printf("using cert-file %s\n", certFile)
    log.Printf("using key-file %s\n", keyFile)
    log.Printf("Serving TLS")
    log.Fatal(http.ListenAndServeTLS(*addr, certFile, keyFile, mux))
  }

  log.Fatal(http.ListenAndServe(*addr, mux))
}

