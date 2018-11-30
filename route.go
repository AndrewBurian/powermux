package powermux

import (
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
)

const (
	methodAny = "ANY"
	notFound  = "NOT_FOUND"
)

type childList []*Route

func (l childList) Len() int {
	return len(l)
}

func (l childList) Less(i, j int) bool {
	return l[i].pattern < l[j].pattern
}

func (l childList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func (l childList) Search(pattern string) *Route {
	index := sort.Search(l.Len(), func(i int) bool {
		return l[i].pattern >= pattern
	})

	if index < l.Len() && l[index].pattern == pattern {
		return l[index]
	}

	return nil
}

type verbFlag uint8

const (
	flagGet = verbFlag(1) << iota
	flagHead
	flagPost
	flagPut
	flagPatch
	flagDelete
	flagConnect
	flagOptions
	flagAny = ^verbFlag(0)
)

// Check if the verb matches the available flags
// never match a zero flag
func (f verbFlag) Matches(v verbFlag) bool {
	if v == 0 {
		return false
	}
	return f&v == v
}

func getVerbFlag(verb string) verbFlag {
	switch verb {
	case http.MethodGet:
		return flagGet
	case http.MethodHead:
		return flagHead
	case http.MethodPost:
		return flagPost
	case http.MethodPut:
		return flagPut
	case http.MethodPatch:
		return flagPatch
	case http.MethodDelete:
		return flagDelete
	case http.MethodConnect:
		return flagConnect
	case http.MethodOptions:
		return flagOptions
	default:
		panic("powermux: getVerbFlag: not a valid http method: " + verb)
	}
}

type middlewareVerb struct {
	mid  Middleware
	verb verbFlag
}

var pathPartsPool = &sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 5)
	},
}

// A Route represents a specific path for a request.
// Routes can be absolute paths, rooted subtrees, or path parameters that accept any stringRoutes.
type Route struct {
	// the pattern our node matches
	pattern string
	// the full path to this node
	fullPath string
	// if we are a named path param node '/:name'
	isParam bool
	// the name of our path parameter
	paramName string
	// if we are a rooted sub tree '/dir/*'
	isWildcard bool
	// the array of middleware this node invokes
	middleware []*middlewareVerb
	// child nodes
	children childList
	// child node for path parameters
	paramChild *Route
	// set if there's a wildcard handler child (lowest priority)
	wildcardChild *Route
	// the map of handlers for different methods
	handlers map[string]http.Handler
}

// newRoute allocates all the structures required for a route node.
// Default pattern is "" which matches only the top level node.
func newRoute() *Route {
	return &Route{
		handlers:   make(map[string]http.Handler),
		middleware: make([]*middlewareVerb, 0),
		children:   make([]*Route, 0),
	}
}

// execute sets up the tree traversal required to get the execution instructions for
// a route.
func (r *Route) execute(ex *routeExecution, method, pattern string) {

	pathParts := pathPartsPool.Get().([]string)[0:0]
	pathParts = append(pathParts, "")
	start := 1
	for i := 1; i < len(pattern); i++ {
		if pattern[i] == '/' {
			pathParts = append(pathParts, pattern[start:i])
			i++
			start = i
		}
	}

	// get the trailing path param
	if pattern != "/" {
		pathParts = append(pathParts, pattern[start:])
	}

	// Fill the execution
	r.getExecution(method, pathParts, ex)

	// return path parts
	pathPartsPool.Put(pathParts)
}

// getExecution is a recursive step in the tree traversal. It checks to see if this node matches,
// fills out any instructions in the execution, and returns. The return value indicates only if
// this node matched, not if anything was added to the execution.
func (r *Route) getExecution(method string, pathParts []string, ex *routeExecution) {

	curRoute := r
	verb := getVerbFlag(method)

	for {

		// save all the middleware for matching verbs
		for i := range curRoute.middleware {
			if curRoute.middleware[i].verb.Matches(verb) {
				ex.middleware = append(ex.middleware, curRoute.middleware[i].mid)
			}
		}

		// save not found handler
		if h, ok := curRoute.handlers[notFound]; ok {
			ex.notFound = h
		}

		// save options handler
		if method == http.MethodOptions {
			if h, ok := curRoute.handlers[http.MethodOptions]; ok {
				ex.handler = h
			}
		}

		// save path parameters
		if curRoute.isParam {
			// Errors here will never happen as Go's http server sanitizes inputs before
			// they are handled by the mux, therefore the error return is ignored
			value, _ := url.PathUnescape(pathParts[0])
			ex.params[curRoute.paramName] = value
		}

		// check if this is the bottom of the path
		if len(pathParts) == 1 || curRoute.isWildcard {

			// hit the bottom of the tree, see if we have a handler to offer
			curRoute.getHandler(method, ex)

			if curRoute.fullPath == "" {
				ex.pattern = "/"
			} else {
				ex.pattern = curRoute.fullPath
			}
			return

		}

		// iterate over our children looking for deeper to go

		// binary search over regular children
		if child := curRoute.children.Search(pathParts[1]); child != nil {
			pathParts = pathParts[1:]
			curRoute = child
			continue
		}

		// try for params and wildcard children
		if curRoute.paramChild != nil {
			pathParts = pathParts[1:]
			curRoute = curRoute.paramChild
			continue
		}
		if curRoute.wildcardChild != nil {
			pathParts = pathParts[1:]
			curRoute = curRoute.wildcardChild
			continue
		}

		return
	}
}

// getHandler is a convenience function for choosing a handler from the route's map of options
// Order of precedence:
// 1. An exact method match
// 2. HEAD requests can use GET handlers
// 3. The ANY handler
// 4. A generated Options handler if this is an options request and no previous handler is set
// 5. A generated Method Not Allowed response
func (r *Route) getHandler(method string, ex *routeExecution) {
	// check specific method match
	if h, ok := r.handlers[method]; ok {
		ex.handler = h
		return
	}

	// if this is a HEAD we can fall back on GET
	if method == http.MethodHead {
		if h, ok := r.handlers[http.MethodGet]; ok {
			ex.handler = h
			return
		}
	}

	// check the ANY handler
	if h, ok := r.handlers[methodAny]; ok {
		ex.handler = h
		return
	}

	// last ditch effort is to generate our own method not allowed handler
	// this is regenerated each time in case routes are added during runtime
	// not generated if a previous handler is already set
	if ex.handler == nil {
		ex.handler = r.methodNotAllowed()
	}
	return
}

// Route walks down the route tree following pattern and returns either a new or previously
// existing node that represents that specific path.
func (r *Route) Route(path string) *Route {

	// prepend a leading slash if not present
	if path[0] != '/' {
		path = "/" + path
	}

	// remove the tailing slash if it is present
	if path != "/" {
		path = strings.TrimRight(path, "/")
	}

	// append our node name to the search if we're not root
	if r.pattern != "" {
		path = r.pattern + path
	}

	// chop the path into pieces
	pathParts := strings.Split(path, "/")

	// handle the case where we're the root node
	if path == "/" {
		// strings.Split will have returned ["", ""]
		// drop one of them
		pathParts = pathParts[1:]
	}

	// find/create the new path
	return r.create(pathParts, r.fullPath)
}

// Create descends the tree following path, creating nodes as needed and returns the target node
func (r *Route) create(path []string, parentPath string) *Route {

	// ensure this path matches us
	if r.pattern != path[0] {
		// not us
		return nil
	}

	// if this is us, return, no creation necessary
	if len(path) == 1 {
		return r
	}

	// iterate over all children looking for a place to put this
	for _, child := range r.getChildren() {
		if r := child.create(path[1:], r.fullPath); r != nil {
			return r
		}
	}

	// child can't create it, so we will
	newRoute := newRoute()

	// set the pattern name
	newRoute.pattern = path[1]
	newRoute.fullPath = r.fullPath + "/" + path[1]

	// check if it's a path param
	if strings.HasPrefix(path[1], ":") {
		newRoute.isParam = true
		newRoute.paramName = strings.TrimLeft(path[1], ":")

		// save it in the correct place
		r.paramChild = newRoute

	} else if path[1] == "*" {
		// check if this is a rooted subtree
		newRoute.isWildcard = true

		// save to wildcard child
		r.wildcardChild = newRoute

		// go no deeper
		return newRoute
	} else {
		// Just a regular child
		r.children = append(r.children, newRoute)

		// sort children alphabetically for efficient run time searching
		sort.Sort(r.children)
	}

	// the cycle continues
	return newRoute.create(path[1:], r.fullPath)
}

// stringRoutes returns the stringRoutes representation of this route and all below it.
func (r *Route) stringRoutes(routes *[]string) {

	var thisRoute string

	// handle root node
	if r.fullPath == "" {
		thisRoute = "/"
	} else {
		thisRoute = r.fullPath
	}

	if len(r.handlers) > 0 {
		thisRoute = thisRoute + "\t["
		methods := make([]string, 0, 8)
		for method := range r.handlers {
			methods = append(methods, method)
		}
		thisRoute = thisRoute + strings.Join(methods, ", ") + "]"
		*routes = append(*routes, thisRoute)
	}

	// recursion
	for _, child := range r.getChildren() {
		child.stringRoutes(routes)
	}
}

// getChildren returns all the routes with the correct order of precedence
func (r *Route) getChildren() []*Route {

	// allocate once
	allRoutes := make([]*Route, 0, len(r.children)+2)

	// start with the normal routes
	allRoutes = append(allRoutes, r.children...)

	// then add the param child
	if r.paramChild != nil {
		allRoutes = append(allRoutes, r.paramChild)
	}

	// then add the wildcard child
	if r.wildcardChild != nil {
		allRoutes = append(allRoutes, r.wildcardChild)
	}

	return allRoutes
}

// Middleware adds a middleware to this Route.
//
// Middlewares are executed if the path to the target route crosses this route.
func (r *Route) Middleware(m Middleware) *Route {
	r.middleware = append(r.middleware, &middlewareVerb{
		mid:  m,
		verb: flagAny,
	})
	return r
}

// MiddlewareFor adds a middleware to this node, but will only be executed
// for requests with the verb specified.
// Verbs are case sensitive, and should use the `http.Method*` constants.
// Panics if any of the verbs provided are unknown.
func (r *Route) MiddlewareFor(m Middleware, verbs ...string) *Route {

	// Equivalent to none
	if len(verbs) == 0 {
		return r
	}

	f := verbFlag(0)
	for _, verb := range verbs {
		f = f | getVerbFlag(verb)
	}

	// we don't check if this is equivalent to flagAny since a
	// fully loaded flag set is the same as the flagAny

	r.middleware = append(r.middleware, &middlewareVerb{
		mid:  m,
		verb: f,
	})

	return r

}

// MiddlewareExceptFor adds a middleware to this node, but will only be executed
// for requests that are not in the list of verbs.
// Verbs are case sensitive, and should use the `http.Method*` constants.
// Panics if any of the verbs provided are unknown.
func (r *Route) MiddlewareExceptFor(m Middleware, verbs ...string) *Route {

	// Equivalent to any
	if len(verbs) == 0 {
		return r.Middleware(m)
	}

	// build the list as if we are calculating For
	f := verbFlag(0)
	for _, verb := range verbs {
		f = f | getVerbFlag(verb)
	}

	// then invert to get ExceptFor
	f = ^f

	// Equivalent to none
	if f == 0 {
		return r
	}

	r.middleware = append(r.middleware, &middlewareVerb{
		mid:  m,
		verb: f,
	})

	return r

}

// MiddlewareExceptForOptions is shorthand for MiddlewareExceptFor with
// http.MethodOptions as the only excepted method
func (r *Route) MiddlewareExceptForOptions(m Middleware) *Route {
	return r.MiddlewareExceptFor(m, http.MethodOptions)
}

// MiddlewareFunc registers a plain function as a middleware.
func (r *Route) MiddlewareFunc(m MiddlewareFunc) *Route {
	return r.Middleware(MiddlewareFunc(m))
}

// Any registers a catch-all handler for any method sent to this route.
// This takes lower precedence than a specific method match.
func (r *Route) Any(handler http.Handler) *Route {
	r.handlers[methodAny] = handler
	return r
}

// AnyFunc registers a plain function as a catch-all handler
// for any method sent to this route.
// This takes lower precedence than a specific method match.
func (r *Route) AnyFunc(f http.HandlerFunc) *Route {
	return r.Any(http.HandlerFunc(f))
}

// Post adds a handler for POST methods to this route.
func (r *Route) Post(handler http.Handler) *Route {
	r.handlers[http.MethodPost] = handler
	return r
}

// PostFunc adds a plain function as a handler
// for POST methods to this route.
func (r *Route) PostFunc(f http.HandlerFunc) *Route {
	return r.Post(http.HandlerFunc(f))
}

// Put adds a handler for PUT methods to this route.
func (r *Route) Put(handler http.Handler) *Route {
	r.handlers[http.MethodPut] = handler
	return r
}

// PutFunc adds a plain function as a handler
// for PUT methods to this route.
func (r *Route) PutFunc(f http.HandlerFunc) *Route {
	return r.Put(http.HandlerFunc(f))
}

// Patch adds a handler for PATCH methods to this route.
func (r *Route) Patch(handler http.Handler) *Route {
	r.handlers[http.MethodPatch] = handler
	return r
}

// PatchFunc adds a plain function as a handler
// for PATCH methods to this route.
func (r *Route) PatchFunc(f http.HandlerFunc) *Route {
	return r.Patch(http.HandlerFunc(f))
}

// Get adds a handler for GET methods to this route.
// GET handlers will also be called for HEAD requests
// if no specific HEAD handler is registered.
func (r *Route) Get(handler http.Handler) *Route {
	r.handlers[http.MethodGet] = handler
	return r
}

// GetFunc adds a plain function as a handler
// for GET methods to this route.
// GET handlers will also be called for HEAD requests
// if no specific HEAD handler is registered.
func (r *Route) GetFunc(f http.HandlerFunc) *Route {
	return r.Get(http.HandlerFunc(f))
}

// Delete adds a handler for DELETE methods to this route.
func (r *Route) Delete(handler http.Handler) *Route {
	r.handlers[http.MethodDelete] = handler
	return r
}

// DeleteFunc adds a plain function as a handler
// for DELETE methods to this route.
func (r *Route) DeleteFunc(f http.HandlerFunc) *Route {
	return r.Delete(http.HandlerFunc(f))
}

// Head adds a handler for HEAD methods to this route.
func (r *Route) Head(handler http.Handler) *Route {
	r.handlers[http.MethodHead] = handler
	return r
}

// HeadFunc adds a plain function as a handler
// for HEAD methods to this route.
func (r *Route) HeadFunc(f http.HandlerFunc) *Route {
	return r.Head(http.HandlerFunc(f))
}

// Connect adds a handler for CONNECT methods to this route.
func (r *Route) Connect(handler http.Handler) *Route {
	r.handlers[http.MethodConnect] = handler
	return r
}

// ConnectFunc adds a plain function as a handler
// for CONNECT methods to this route.
func (r *Route) ConnectFunc(f http.HandlerFunc) *Route {
	return r.Connect(http.HandlerFunc(f))
}

// Options adds a handler for OPTIONS methods to this route.
// This handler will also be called for any routes further down the path
// from this point if no other OPTIONS handlers are registered below.
func (r *Route) Options(handler http.Handler) *Route {
	r.handlers[http.MethodOptions] = handler
	return r
}

// OptionsFunc adds a plain function as a handler
// for OPTIONS methods to this route.
// This handler will also be called for any routes further down the path
// from this point if no other OPTIONS handlers are registered below.
func (r *Route) OptionsFunc(f http.HandlerFunc) *Route {
	return r.Options(http.HandlerFunc(f))
}

// NotFound adds a handler for requests that do not correspond to a route.
// This handler will also be called for any routes further down the path
// from this point if no other not found handlers are registered below.
func (r *Route) NotFound(handler http.Handler) *Route {
	r.handlers[notFound] = handler
	return r
}

// NotFoundFunc adds a plain function as a handler for requests
// that do not correspond to a route.
// This handler will also be called for any routes further down the path
// from this point if no other not found handlers are registered below.
func (r *Route) NotFoundFunc(f http.HandlerFunc) *Route {
	return r.NotFound(http.HandlerFunc(f))
}
