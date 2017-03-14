package powermux

import "net/http"

// ServerMux is the multiplexer for http requests
type ServeMux struct {
	NotFound http.Handler
	baseRoute *Route
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		baseRoute: &Route{
			pattern: "/",
		},
	}
}

func (s *ServeMux) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	// Get the handler for this request
	handler, _, _ := s.HandlerAndMiddleware(req)

	// If there is no handler, run the not found handler
	if handler == nil {
		s.NotFound.ServeHTTP(rw, req)
		return
	}

	// run all the middleware sequentially
	//for middleware := range middlewares {
		//todo this logic is fucked
	//}

	// run the handler
	handler.ServeHTTP(rw, req)
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
	return nil, nil, ""
}

func (s *ServeMux) Route(pattern string) *Route {
	return nil
}