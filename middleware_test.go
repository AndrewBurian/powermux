package powermux

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware_getNextMiddleware(t *testing.T) {
	m1 := func(w http.ResponseWriter, r *http.Request, n NextMiddlewareFunc) {
		io.WriteString(w, "middleware 1- ")
		n(w, r)
	}

	m2 := func(w http.ResponseWriter, r *http.Request, n NextMiddlewareFunc) {
		io.WriteString(w, "middleware 2- ")
		n(w, r)
	}

	m3 := func(w http.ResponseWriter, r *http.Request, n NextMiddlewareFunc) {
		n(w, r)
		io.WriteString(w, "middleware 3- ")
	}

	h := func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "handler- ")
	}

	mids := make([]Middleware, 0, 3)

	mids = append(mids, MiddlewareFunc(m1))
	mids = append(mids, MiddlewareFunc(m2))
	mids = append(mids, MiddlewareFunc(m3))

	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", bytes.NewBufferString("Hey"))

	f := getNextMiddleware(mids, http.HandlerFunc(h))
	f(recorder, req)

	if recorder.Body.String() != "middleware 1- middleware 2- handler- middleware 3- " {

		t.Error("Middlewares not executed in order")
		t.Log(recorder.Body.String())
	}
}
