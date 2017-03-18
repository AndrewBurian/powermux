package powermux

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRoute_MethodNotAllowed(t *testing.T) {
	// new route
	r := newRoute()

	// add a GET and DELETE handler
	r.Get(http.NotFoundHandler())
	r.Delete(http.NotFoundHandler())

	ex := &routeExecution{}

	// Try for a POST handler
	r.getHandler(http.MethodPost, ex)

	if ex.handler == nil {
		t.Fatal("Nil handler returned")
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("Hello"))

	ex.handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Error("Wrong method returned")
	}

	allows := strings.Split(rec.HeaderMap.Get("Allow"), ", ")
	allowedMethods := make(map[string]bool)
	for _, allow := range allows {
		allowedMethods[allow] = true
	}

	if !allowedMethods[http.MethodGet] || !allowedMethods[http.MethodDelete] {
		t.Error("Did not allow all required methods")
	}
	if len(allowedMethods) > 2 {
		t.Error("Excessive methods allowed")
	}
}
