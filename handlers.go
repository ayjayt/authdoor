package authdoor

import (
	"net/http"
)

type redirectSchemeHandler struct {
	scheme string
	code   int
}

func (rsh *redirectSchemeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	newURL := r.URL
	if r.URL.IsAbs() == false {
		r.URL.Host = r.Host
	}
	newURL.Scheme = rsh.scheme
	http.Redirect(w, r, newURL.String(), rsh.code)
}

func RedirectSchemeHandler(scheme string, code int) http.Handler {
	return &redirectSchemeHandler{scheme, code}
}

type ReverseProxy struct {
	http.Handler
}

func NewSingleHostReverseProxy(target string) (*ReverseProxy, error) {
	targetURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	return &ReverseProxy{
		Handler: httputil.NewSingleHostReverseProxy(targetURL),
	}, nil
}
