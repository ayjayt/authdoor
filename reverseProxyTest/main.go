package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"

	"github.com/ayjayt/authdoor"
)

var proxyPort = flag.String("proxy", "9010", "port number for the proxy")
var backendPort = flag.String("backend", "9011", "port number for the backend")

func init() {
	flag.Parse()
}

// DumperHandler just dumps the request before calling the actual handler
type DumperHandler struct {
	http.Handler
}

// NewDumper is just a constructor for a dumper handler
func NewDumper(h http.Handler) http.Handler {
	return &DumperHandler{Handler: h}
}

// ServeHTTP is the method allowing us to implement http.Handler interface with DumperHandler and also dump the request
func (dh *DumperHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Dump:\n%q\n", dump)
	dh.Handler.ServeHTTP(w, r)
}

// OkHandler is an http.Handler that just returns back ok
type OkHandler struct {
}

// ServeHTTP is the method on OkHandler that fuffils the http.Handler method
func (oh *OkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("\n"))
}

func main() {
	backend := NewDumper(&OkHandler{})
	proxy, _ := authdoor.NewSingleHostReverseProxy("localhost:" + *backendPort)
	proxyWrapped := NewDumper(proxy)
	go http.ListenAndServe(*backendPort, backend)
	http.ListenAndServe(*proxyPort, proxyWrapped)
}
