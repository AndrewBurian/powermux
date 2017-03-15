package powermux

import (
	"net/http"
	"strings"
)

const (
	methodAny = "ANY"
	notFound  = "NOT_FOUND"
)

// routeExecution is the complete instructions for running serve on a route
type routeExecution struct {
	pattern    []string
	params     map[string]string
	notFound   http.Handler
	middleware []Middleware
	handler    http.Handler
}

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

// newRoute allocates all the structures required for a route node
// default pattern is "/"
func newRoute() *Route {
	return &Route{
		pattern:    "/",
		handlers:   make(map[string]http.Handler),
		middleware: make([]Middleware, 0),
		children:   make([]*Route, 0),
	}
}

func (r *Route) execute(method, pattern string) *routeExecution {

	pathParts := strings.Split(pattern, "/")

	// Create a new routeExecution
	ex := &routeExecution{
		pattern:    make([]string, 0, len(pathParts)),
		middleware: make([]Middleware, 0),
		params:     make(map[string]string),
	}

	// Fill the execution
	r.getExecution(method, pathParts, ex)

	// return the result
	return ex
}

func (r *Route) getExecution(method string, pathParts []string, ex *routeExecution) bool {

	var match bool

	// If this node is a path parameter, we match any string
	if r.isParam {
		// save the path parameter
		ex.params[r.paramName] = pathParts[0]
		match = true
	}

	// check if we're a direct match
	if r.pattern == pathParts[0] {
		match = true
	}

	// if we don't match, return
	if !match {
		return false
	}

	// save this node as part of the path
	ex.pattern = append(ex.pattern, r.pattern)

	// save all the middleware
	ex.middleware = append(ex.middleware, r.middleware...)

	// save not found handler
	if h, ok := r.handlers[notFound]; ok {
		ex.notFound = h
	}

	// save options handler
	if method == http.MethodOptions {
		if h, ok := r.handlers[http.MethodOptions]; ok {
			ex.handler = h
		}
	}

	// check if this is the bottom of the path
	if len(pathParts) == 1 {

		// hit the bottom of the tree, see if we have a handler to offer
		ex.handler = r.getHandler(method)
		return true

	}

	// iterate over our children looking for deeper to go
	for _, child := range r.children {
		if found := child.getExecution(method, pathParts[1:], ex); found {
			return found
		}
	}

	// if we're a rooted subtree, we can still use our handler
	if r.isRoot {
		ex.handler = r.getHandler(method)
	}

	return true
}

// getHandler is a convenience function for choosing a handler from the route's map of options
// Order of precedence:
// 1. An exact method match
// 2. HEAD requests can use GET handlers
// 3. The ANY handler
// 4. A generated Method Not Allowed response
func (r *Route) getHandler(method string) http.Handler {
	// check specific method match
	if h, ok := r.handlers[method]; ok {
		return h
	}

	// if this is a HEAD we can fall back on GET
	if method == http.MethodHead {
		if h, ok := r.handlers[http.MethodGet]; ok {
			return h
		}
	}

	// check the ANY handler
	if h, ok := r.handlers[methodAny]; ok {
		return h
	}

	// last ditch effort is to generate our own method not allowed handler
	// this is regenerated each time in case routes are added during runtime
	return r.methodNotAllowed()
}

// Route walks down the route tree following pattern
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

// Create descends the tree following path, creating nodes as needed and returns the target node
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
	newRoute := newRoute()

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
