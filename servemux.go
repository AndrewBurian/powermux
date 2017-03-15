package powermux

import (
	"context"
	"net/http"
	"strings"
)

// ServerMux is the multiplexer for http requests
type ServeMux struct {
	NotFound  http.Handler
	baseRoute *Route
}

// ctxKey is the key type used for path parameters in the request context
type ctxKey string

// GetPathParams gets named path parameters and their values from the request
//
// the path '/users/:name' given '/users/andrew' will have `GetPathParams(r, "name")` => `"andrew"`
// unset values return an empty string
func GetPathParam(req *http.Request, name string) (value string) {
	name, _ = req.Context().Value(ctxKey(name)).(string)
	return
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		baseRoute: &Route{
			pattern: "/",
		},
	}
}

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

func (s *ServeMux) Handle(pattern string, handler http.Handler) {
	s.Route(pattern).Any(handler)
}

func (s *ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	s.Handle(pattern, http.HandlerFunc(handler))
}

func (s *ServeMux) Handler(r *http.Request) (http.Handler, string) {
	handler, _, pattern := s.HandlerAndMiddleware(r)
	return handler, pattern
}

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

func (s *ServeMux) Route(pattern string) *Route {
	return s.baseRoute.Route(pattern)
}
