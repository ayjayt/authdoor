package authdoor

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

// redirectSchemeHandler holds the new scheme (usually HTTPS) and redirect code (30x) to use during redirection
type redirectSchemeHandler struct {
	scheme string
	code   int
}

// ServeHTTP rewrites the requests URL and appropriately and then calls Redirect.
func (rsh *redirectSchemeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newURL := r.URL
	if r.URL.IsAbs() == false {
		r.URL.Host = r.Host
	}
	newURL.Scheme = rsh.scheme
	http.Redirect(w, r, newURL.String(), rsh.code)
}

// RedirectSchemeHandler returns a new http.Handler
func RedirectSchemeHandler(scheme string, code int) http.Handler {
	defaultLogger.Info("Creating new scheme redirect to " + scheme + " with code " + strconv.Itoa(code))
	return &redirectSchemeHandler{scheme, code}
}

// ReverseProxy is a wrapper for httputil's reverse proxy that just does the tedious working of parsing a url string. Paths are always passed to the proxy, it's just the host that's rewritten. I'm not sure if the proxy receives the original host.
type ReverseProxy struct {
	http.Handler
}

// NewSingleHostReverseProxy is the constructor for the ReverseProxy struct that actually does the work
func NewSingleHostReverseProxy(target string) (*ReverseProxy, error) {
	defaultLogger.Info("Creating new single host reverse proxy to " + target)
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	return &ReverseProxy{
		Handler: httputil.NewSingleHostReverseProxy(targetURL),
	}, nil
}
