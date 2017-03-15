package powermux

import (
	"context"
	"net/http"
	"strings"
)

// ServeMux is the multiplexer for http requests
type ServeMux struct {
	baseRoute *Route
}

// ctxKey is the key type used for path parameters in the request context
type ctxKey string

// GetPathParam gets named path parameters and their values from the request
//
// the path '/users/:name' given '/users/andrew' will have `GetPathParams(r, "name")` => `"andrew"`
// unset values return an empty string
func GetPathParam(req *http.Request, name string) (value string) {
	name, _ = req.Context().Value(ctxKey(name)).(string)
	return
}

// NewServeMux creates a new multiplexer, and sets up a default not found handler
func NewServeMux() *ServeMux {
	s := &ServeMux{
		baseRoute: newRoute(),
	}
	s.NotFound(http.NotFoundHandler())
	return s
}

// ServeHTTP dispatches the request to the handler whose pattern most closely matches the request URL.
func (s *ServeMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Get the route execution
	ex := s.baseRoute.execute(req.Method, req.URL.Path)

	// If there is no handler, run the not found handler
	if ex.handler == nil {
		ex.handler = ex.notFound
	}

	// set all the path params
	if len(ex.params) > 0 {
		var ctx context.Context
		for key, val := range ex.params {
			ctx = context.WithValue(req.Context(), ctxKey(key), val)
		}
		req = req.WithContext(ctx)
	}

	// Run a middleware/handler closure to nest all middleware
	f := getNextMiddleware(ex.middleware, ex.handler)
	f(rw, req)
}

// Handle registers the handler for the given pattern. If a handler already exists for pattern it is overwritten.
func (s *ServeMux) Handle(pattern string, handler http.Handler) {
	s.Route(pattern).Any(handler)
}

// Handle adds middleware for the given pattern.
func (s *ServeMux) Middleware(pattern string, middleware Middleware) {
	s.Route(pattern).Middleware(middleware)
}

// HandleFunc registers the handler function for the given pattern.
func (s *ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.Handle(pattern, http.HandlerFunc(handler))
}

// Handler returns the handler to use for the given request, consulting r.Method, r.Host, and r.URL.Path.
// It always returns a non-nil handler. If the path is not in its canonical form, the handler will be an
// internally-generated handler that redirects to the canonical path.
//
// Handler also returns the registered pattern that matches the request or, in the case of internally-generated
// redirects, the pattern that will match after following the redirect.
//
// If there is no registered handler that applies to the request, Handler returns a “page not found” handler
// and an empty pattern.
func (s *ServeMux) Handler(r *http.Request) (http.Handler, string) {
	handler, _, pattern := s.HandlerAndMiddleware(r)
	return handler, pattern
}

// HandlerAndMiddleware returns the same as Handler, but with the addition of an array of middleware, in the order
// they would have been executed
func (s *ServeMux) HandlerAndMiddleware(r *http.Request) (http.Handler, []Middleware, string) {

	// Get the route execution
	ex := s.baseRoute.execute(r.Method, r.URL.Path)

	// reconstruct the path
	pattern := strings.Join(ex.pattern, "/")

	// fall back on not found handler if necessary
	if ex.handler == nil {
		ex.handler = ex.notFound
	}

	return ex.handler, ex.middleware, pattern
}

// Route returns the route from the root of the domain to the given pattern
func (s *ServeMux) Route(pattern string) *Route {
	return s.baseRoute.Route(pattern)
}

// NotFound sets the default not found handler for the server
func (s *ServeMux) NotFound(handler http.Handler) {
	s.baseRoute.NotFound(handler)
}
