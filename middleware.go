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

// Middleware handles HTTP requests and optionally passes them along to the next handler in the chain
type Middleware interface {
	ServeHTTPMiddleware(http.ResponseWriter, *http.Request, NextMiddlewareFunc)
}

// getNextMiddleware returns the first middleware of a recursive closure.
// The returned middleware will have the next middleware in the array available to it as a parameter
// and the last middleware will have the final handler
func getNextMiddleware(mids []Middleware, h http.Handler) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if len(mids) > 0 {
			mids[0].ServeHTTPMiddleware(w, r, getNextMiddleware(mids[1:], h))
		} else {
			h.ServeHTTP(w, r)
		}
	}
}
