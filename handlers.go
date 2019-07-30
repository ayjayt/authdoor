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
	newURL.Scheme = rsh.scheme
	http.Redirect(w, r, newURL.String(), rsh.code)
}

func RedirectSchemeHandler(scheme string, code int) http.Handler {
	return &redirectSchemeHandler{scheme, code}
}
