package powermux

import (
	"net/http"
	"strings"
)

const methodAny = "ANY"
const notFound = "NOT_FOUND"

type Route struct {
	// the pattern our node matches
	pattern string
	// if we are a named path param node '/:name'
	isParam bool
	// the name of our path parameter
	paramName string
	// if we are a rooted sub tree '/dir/'
	isRoot bool
	// the array of middleware this node invokes
	middleware []Middleware
	// child nodes
	children []*Route
	// the map of handlers for different methods
	handlers map[string]http.Handler
}

func (r *Route) get(method, pattern string) (http.Handler, []Middleware, map[string]string) {
	middleware := make([]Middleware, 0)
	pathParts := strings.Split(pattern, "/")
	pathParams := make(map[string]string)

	handler := r.findAll(method, pathParts, &middleware, pathParams)

	return handler, middleware, pathParams
}

func (r *Route) findAll(method string, pathParts []string, mid *[]Middleware, params map[string]string) http.Handler {

	var match bool

	// If this node is a path parameter, we match any string
	if r.isParam {
		// save the path parameter
		params[r.paramName] = pathParts[0]
		match = true
	}

	// check if we're a direct match
	if r.pattern == pathParts[0] {
		match = true
	}

	// if we don't match, return
	if !match {
		return nil
	}

	// first save all the middleware
	*mid = append(*mid, r.middleware...)

	// ensure this is not the bottom of the path
	if len(pathParts) == 1 {

		// hit the bottom of the tree, see if we have a handler to offer
		if h, ok := r.handlers[method]; ok {
			return h
		}
		if h, ok := r.handlers[methodAny]; ok {
			return h
		}

		// end of the line, no handlers found
		return nil

	}

	// iterate over our children looking for deeper to go
	for _, child := range r.children {
		if h := child.findAll(method, pathParts[1:], mid, params); h != nil {
			return h
		}
	}

	// if we're a rooted subtree, we can still return
	if r.isRoot {
		if h, ok := r.handlers[method]; ok {
			return h
		}
		if h, ok := r.handlers[methodAny]; ok {
			return h
		}
		// no method found for this subtree
		return nil
	}

	// children have nothing to offer and we are not the target
	return nil
}

// Creates and finds all in one go
func (r *Route) Route(pattern string) *Route {

	pattern = strings.TrimLeft(pattern, "/")

	// remove leading and trailing slash
	trimPattern := strings.Trim(pattern, "/")

	// break it up into pieces
	parts := strings.Split(trimPattern, "/")

	// if the path ended with a slash, indicating a subtree, re add it
	if strings.HasSuffix(pattern, "/") {
		parts = append(parts, "/")
	}

	path := make([]string, 0, len(parts)+1)
	// add the leading slash
	path = append(path, "/")

	if parts[0] != "" {
		path = append(path, parts...)
	}

	/*
		/     => ['/']
		/a    => ['/', 'a']
		/a/b  => ['/', 'a', 'b']
		/a/b/ => ['/', 'a', 'b', '/']
	*/

	// find/create the new path
	return r.create(path)
}

func (r *Route) create(path []string) *Route {

	// ensure this path matches us
	if !r.isParam && r.pattern != path[0] {
		// not us
		return nil
	}

	// if this is us, return, no creation necessary
	if len(path) == 1 {
		return r
	}

	// iterate over all children looking for a place to put this
	for _, child := range r.children {
		if r := child.create(path[1:]); r != nil {
			return r
		}
	}

	// child can't create it, so we will
	newRoute := &Route{
		handlers: make(map[string]http.Handler),
	}

	// save child
	r.children = append(r.children, newRoute)

	// check if it's a path param
	if strings.HasPrefix(path[1], ":") {
		newRoute.isParam = true
		newRoute.paramName = strings.TrimLeft(path[1], ":")
	} else {
		newRoute.pattern = path[1]
	}

	// check if this is a rooted subtree
	if path[1] == "/" {
		// go no deeper
		newRoute.isRoot = true
		return newRoute
	}

	// the cycle continues
	return newRoute.create(path[1:])
}

//
func (r *Route) Middleware(m Middleware) *Route {
	r.middleware = append(r.middleware, m)
	return r
}

// MiddlewareFunc registers a plain function as
func (r *Route) MiddlewareFunc(m MiddlewareFunc) *Route {
	return r.Middleware(MiddlewareFunc(m))
}

func (r *Route) Any(handler http.Handler) *Route {
	r.handlers[methodAny] = handler
	return r
}

func (r *Route) Post(handler http.Handler) *Route {
	r.handlers[http.MethodPost] = handler
	return r
}

func (r *Route) Patch(handler http.Handler) *Route {
	r.handlers[http.MethodPatch] = handler
	return r
}

func (r *Route) Get(handler http.Handler) *Route {
	r.handlers[http.MethodGet] = handler
	return r
}

func (r *Route) Delete(handler http.Handler) *Route {
	r.handlers[http.MethodDelete] = handler
	return r
}

func (r *Route) Head(handler http.Handler) *Route {
	r.handlers[http.MethodHead] = handler
	return r
}

func (r *Route) Options(handler http.Handler) *Route {
	r.handlers[http.MethodOptions] = handler
	return r
}

func (r *Route) Connect(handler http.Handler) *Route {
	r.handlers[http.MethodConnect] = handler
	return r
}

func (r *Route) NotFound(handler http.Handler) *Route {
	r.handlers[notFound] = handler
	return r
}
