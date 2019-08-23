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
	me string
	http.Handler
}

// NewDumper is just a constructor for a dumper handler
func NewDumper(h http.Handler, me string) http.Handler {
	return &DumperHandler{Handler: h, me: me}
}

// ServeHTTP is the method allowing us to implement http.Handler interface with DumperHandler and also dump the request
func (dh *DumperHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Info:\n%q\n", dh.me)
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
	backend := NewDumper(&OkHandler{}, "backend")
	proxy, _ := authdoor.NewSingleHostReverseProxy("http://localhost:" + *backendPort)
	proxyWrapped := NewDumper(proxy, "proxy")
	go http.ListenAndServe(":"+*backendPort, backend)
	http.ListenAndServe(":"+*proxyPort, proxyWrapped)
	fmt.Printf("Done\n")
}
