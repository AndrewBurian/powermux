package powermux

import (
	"net/http"
)

type NextMiddlewareFunc func(http.ResponseWriter, *http.Request)

//The MiddlewareFunc type is an adapter to allow the use of ordinary functions as HTTP middlewares.
// If f is a function with the appropriate signature, HandlerFunc(f) is a Handler that calls f.
type MiddlewareFunc func(http.ResponseWriter, *http.Request, NextMiddlewareFunc)

//ServeHTTPMiddleware calls f(w, r, n).
func (m MiddlewareFunc) ServeHTTPMiddleware(rw http.ResponseWriter, req *http.Request, n NextMiddlewareFunc) {
	m(rw, req, n)
}

type Middleware interface {
	ServeHTTPMiddleware(http.ResponseWriter, *http.Request, NextMiddlewareFunc)
}

func getNextMiddleware(mids []Middleware, h http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(mids) > 0 {
			mids[0].ServeHTTPMiddleware(w, r, getNextMiddleware(mids[1:], h))
		} else {
			h.ServeHTTP(w, r)
		}
	}
}