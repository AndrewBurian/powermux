package powermux

import (
	"testing"
	"net/http"
	"io"
	"net/http/httptest"
	"bytes"
)

func TestMiddleware_getNextMiddleware(t *testing.T) {
	m1 := func(w http.ResponseWriter, r *http.Request, n NextMiddlewareFunc) {
		io.WriteString(w, "middleware 1\n")
		n(w,r)
	}

	m2 := func(w http.ResponseWriter, r *http.Request, n NextMiddlewareFunc) {
		io.WriteString(w, "middleware 2\n")
		n(w,r)
	}

	h := func (w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "handler")
	}

	mids := make([]Middleware, 0, 2)

	mids = append(mids, MiddlewareFunc(m1))
	mids = append(mids, MiddlewareFunc(m2))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", bytes.NewBufferString("Hey"))

	f := getNextMiddleware(mids, http.HandlerFunc(h))
	f(recorder, req)

	if recorder.Body.String() != "middleware 1\nmiddleware 2\nhandler" {
		t.Error("Middlewares not executed in order")
		t.Log(recorder.Body.String())
	}
}