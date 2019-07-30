package authdoor

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectSchemeHandler(t *testing.T) {
	dut := RedirectSchemeHandler("https", http.StatusMovedPermanently)
	require.IsType(t, &redirectSchemeHandler{}, dut)
	require.Equal(t, "https", dut.(*redirectSchemeHandler).scheme)
	require.Equal(t, http.StatusMovedPermanently, dut.(*redirectSchemeHandler).code)
	req, err := http.NewRequest("GET", "/", nil)
	t.Logf("URL: %v", req.URL.String())
	if err != nil {
		t.Fatal(err)
	}
	responseRecorder := httptest.NewRecorder()
	dut.ServeHTTP(responseRecorder, req)
	require.Equal(t, http.StatusMovedPermanently, responseRecorder.Code)
	require.Equal(t, []string{"https:///"}, responseRecorder.Header()[http.CanonicalHeaderKey("location")])

	req2, err := http.NewRequest("GET", "http://ajpikul.com/thing", nil)
	t.Logf("URL: %v", req2.URL.String())
	if err != nil {
		t.Fatal(err)
	}
	responseRecorder2 := httptest.NewRecorder()
	dut.ServeHTTP(responseRecorder2, req2)
	require.Equal(t, http.StatusMovedPermanently, responseRecorder2.Code)
	require.Equal(t, []string{"https://ajpikul.com/thing"}, responseRecorder2.Header()[http.CanonicalHeaderKey("location")])
	t.Logf("New URL: %+v", responseRecorder2.Header())
}
