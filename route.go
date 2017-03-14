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
	options    http.Handler
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

func (r *Route) execute(method, pattern string) (*routeExecution, bool) {

	pathParts := strings.Split(pattern, "/")

	// Create a new routeExecution
	ex := &routeExecution{
		pattern:    make([]string, 0, len(pathParts)),
		middleware: make([]Middleware, 0),
		params:     make(map[string]string),
	}

	// Fill the execution
	found := r.getExecution(method, pathParts, ex)

	// return the result
	return ex, found
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
		return nil
	}

	// save this node as part of the path
	ex.pattern = append(ex.pattern, r.pattern)

	// first save all the middleware
	ex.middleware = append(ex.middleware, r.middleware...)

	// ensure this is not the bottom of the path
	if len(pathParts) == 1 {

		// hit the bottom of the tree, see if we have a handler to offer
		if h, ok := r.handlers[method]; ok {
			ex.handler = h
			return true
		}
		if h, ok := r.handlers[methodAny]; ok {
			ex.handler = h
			return true
		}

		// end of the line, no handlers found
		return false

	}

	// iterate over our children looking for deeper to go
	for _, child := range r.children {
		if found := child.getExecution(method, pathParts[1:], ex); found {
			return found
		}
	}

	// if we're a rooted subtree, we can still return
	if r.isRoot {
		if h, ok := r.handlers[method]; ok {
			ex.handler = h
			return true
		}
		if h, ok := r.handlers[methodAny]; ok {
			ex.handler = h
			return false
		}
		// no method found for this subtree
		return false
	}

	// children have nothing to offer and we are not the target
	return false
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
