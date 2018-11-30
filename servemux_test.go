package powermux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type dummyHandler string

func (h dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, string(h))
}

func (h dummyHandler) ServeHTTPMiddleware(w http.ResponseWriter, r *http.Request, n func(http.ResponseWriter, *http.Request)) {
	io.WriteString(w, string(h))
	n(w, r)
}

func dummyHandlerFunc(response string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, response)
	}
}

var (
	rightHandler = dummyHandler("right")
	wrongHandler = dummyHandler("wrong")
	mid1         = dummyHandler("mid1")
	mid2         = dummyHandler("mid2")
)

// Ensures that parameter routes have lower precedence than absolute routes
func TestServeMux_ParamPrecedence(t *testing.T) {
	s := NewServeMux()

	s.Route("/users/:id/info").Get(wrongHandler)
	s.Route("/users/jim/info").Get(rightHandler)
	s.Route("/users/:id/detail").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/jim/info", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/users/jim/info" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensures that parameter routes have lower precedence than absolute routes
// and path parameter is properly extracted
func TestServeMux_ParamPrecedenceParamExtraction(t *testing.T) {
	s := NewServeMux()

	var called bool
	var param string

	rightHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		called = true
		param = PathParam(req, "id")
	})

	s.Route("/users/:id/info").Get(wrongHandler)
	s.Route("/users/jim/info").Get(rightHandler)
	s.Route("/users/:id/detail").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/jim/info", nil)
	s.ServeHTTP(nil, req)

	if !called {
		t.Error("None or wrong handler was called")
	}

	if param != "" {
		t.Error("Wrong path param returned")
	}
}

// Ensures that wildcards have the lowest of all precedences
func TestServeMux_WildcardPrecedence(t *testing.T) {
	s := NewServeMux()

	s.Route("/users/*").Get(wrongHandler)
	s.Route("/users/john").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/john", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/users/john" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensures the wildcard handler isn't called when a path param was available
func TestServeMux_WildcardPathPrecedence(t *testing.T) {
	s := NewServeMux()

	s.Route("/users/*").Get(wrongHandler)
	s.Route("/users/:id").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/john", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/users/:id" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensures the wildcard handler isn't called when a path param was available
// and path parameter is properly extracted
func TestServeMux_WildcardPathPrecedenceParamExtraction(t *testing.T) {
	s := NewServeMux()

	var called bool
	var param string

	rightHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		called = true
		param = PathParam(req, "id")
	})

	s.Route("/users/*").Get(wrongHandler)
	s.Route("/users/:id").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/john", nil)
	s.ServeHTTP(nil, req)

	if !called {
		t.Error("None or wrong handler was called")
	}

	if param != "john" {
		t.Error("Wrong path param returned")
	}
}

// Ensures trailing slash redirects are working
func TestServeMux_RedirectSlash(t *testing.T) {
	s := NewServeMux()

	req := httptest.NewRequest(http.MethodGet, "/users/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code != http.StatusPermanentRedirect {
		t.Error("Not redirected")
	}

	if rec.HeaderMap.Get("Location") != "/users" {
		t.Error("Mis-redirected")
	}
}

// Ensures trailing slash redirects are working and return the redirect path
func TestServeMux_RedirectSlashPath(t *testing.T) {
	s := NewServeMux()

	req := httptest.NewRequest(http.MethodGet, "/users/", nil)

	_, path := s.Handler(req)

	if path != "/users" {
		t.Error("Path not corrected")
	}
}

// Ensures we don't redirect the root
func TestServeMux_RedirectRoot(t *testing.T) {
	s := NewServeMux()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Code == http.StatusPermanentRedirect {
		t.Error("Redirected")
	}
}

// Ensure the correct path is matched 1 level
func TestServeMux_HandleCorrectRoute(t *testing.T) {
	s := NewServeMux()

	s.Route("/a").Get(rightHandler)
	s.Route("/b").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/a", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure the correct path is matched at two levels
func TestServeMux_HandleCorrectRouteAfterParam(t *testing.T) {
	s := NewServeMux()

	s.Route("/base/:id/a").Get(rightHandler)
	s.Route("/base/:id/b").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/base/llama/a", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler retured")
	}

	if path != "/base/:id/a" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure the correct path is matched at two levels
// and path parameter is properly extracted
func TestServeMux_HandleCorrectRouteAfterParamExtraction(t *testing.T) {
	s := NewServeMux()

	var called bool
	var param string

	rightHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		called = true
		param = PathParam(req, "id")
	})

	s.Route("/base/:id/a").Get(rightHandler)
	s.Route("/base/:id/b").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/base/llama/a", nil)
	s.ServeHTTP(nil, req)

	if !called {
		t.Error("None or wrong handler was called")
	}

	if param != "llama" {
		t.Error("Wrong path param returned")
	}
}

// Ensure the correct method is matched
func TestServeMux_HandleCorrectMethod(t *testing.T) {
	s := NewServeMux()

	s.Route("/a").Post(rightHandler)
	s.Route("/a").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodPost, "/a", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure the correct method is matched for any
func TestServeMux_HandleCorrectMethodAny(t *testing.T) {
	s := NewServeMux()

	s.Route("/a").Post(wrongHandler)
	s.Route("/a").Get(wrongHandler)
	s.Route("/a").Any(rightHandler)

	req := httptest.NewRequest(http.MethodDelete, "/a", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure the correct method is matched for head
func TestServeMux_HandleCorrectMethodHead(t *testing.T) {
	s := NewServeMux()

	s.Route("/a").Post(wrongHandler)
	s.Route("/a").Get(rightHandler)

	req := httptest.NewRequest(http.MethodHead, "/a", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure a wildcard matches
func TestServeMux_HandleWildcard(t *testing.T) {
	s := NewServeMux()

	s.Route("/a/*").Get(rightHandler)
	s.Route("/b").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/a/llama", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a/*" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure a wildcard matches at depth
func TestServeMux_HandleWildcardDepth(t *testing.T) {
	s := NewServeMux()

	s.Route("/a/*").Get(rightHandler)
	s.Route("/b").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/a/llama/4/5", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a/*" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure order doesn't matter
func TestServeMux_HandleOrder(t *testing.T) {
	s := NewServeMux()

	s.Route("/a").Get(wrongHandler)
	s.Route("/b").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/b", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/b" {
		t.Errorf("Wrong string path: %s", path)
	}
}

func TestServeMux_HandleOptionsAtDepth(t *testing.T) {
	s := NewServeMux()

	s.Route("/a").Options(rightHandler)
	s.Route("/a/b").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodOptions, "/a/b", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a/b" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure routing is not performed on decoded path components
func TestServeMux_EncodedPathComponent(t *testing.T) {
	s := NewServeMux()

	s.Route("/users/:id/info").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/ji%2Fm/info", nil)
	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/users/:id/info" {
		t.Errorf("Wrong string path: %s", path)
	}
}

// Ensure routing is not performed on decoded path components
// and path parameter is properly extracted
func TestServeMux_EncodedPathComponentParamExtraction(t *testing.T) {
	s := NewServeMux()

	var called bool
	var param string

	rightHandler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		called = true
		param = PathParam(req, "id")
	})

	s.Route("/users/:id/info").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/ji%2Fm/info", nil)
	s.ServeHTTP(nil, req)

	if !called {
		t.Error("None or wrong handler was called")
	}

	if param != "ji/m" {
		t.Error("Wrong path param returned")
	}
}

func TestRoute_PermanentRedirect(t *testing.T) {
	s := NewServeMux()

	s.Route("/redir").Redirect("/redirect", true)

	req := httptest.NewRequest(http.MethodGet, "/redir", nil)
	res := httptest.NewRecorder()

	s.ServeHTTP(res, req)

	if res.Code != http.StatusPermanentRedirect {
		t.Error("Should have issued permanemt redirect. Got", res.Code)
	}

	if res.Header().Get("Location") != "/redirect" {
		t.Error("Wrong redirect target. Expected /redirect, got", res.Header().Get("Location"))
	}

}

func TestRoute_TemporaryRedirect(t *testing.T) {
	s := NewServeMux()

	s.Route("/redir").Redirect("/redirect", false)

	req := httptest.NewRequest(http.MethodGet, "/redir", nil)
	res := httptest.NewRecorder()

	s.ServeHTTP(res, req)

	if res.Code != http.StatusTemporaryRedirect {
		t.Error("Should have issued temporary redirect. Got", res.Code)
	}

	if res.Header().Get("Location") != "/redirect" {
		t.Error("Wrong redirect target. Expected /redirect, got", res.Header().Get("Location"))
	}

}

func TestNotFoundEmptyRouteNode(t *testing.T) {
	s := NewServeMux()

	// create but add no handlers
	s.Route("/empty")

	req := httptest.NewRequest(http.MethodGet, "/empty", nil)
	res := httptest.NewRecorder()

	s.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Error("Wrong response code, expected not found, got", res.Code)
	}
}

func TestRoute_Head(t *testing.T) {

	s := NewServeMux()

	s.Route("/").Get(rightHandler)

	req := httptest.NewRequest(http.MethodHead, "/", nil)

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/" {
		t.Error("Wrong path returned", path)
	}

}

func TestRoutePathRoot(t *testing.T) {
	s := NewServeMux()

	s.Route("/").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/" {
		t.Error("Wrong path returned", path)
	}
}

func TestNotFoundFallback(t *testing.T) {
	s := NewServeMux()

	req := httptest.NewRequest(http.MethodGet, "/found", nil)
	res := httptest.NewRecorder()

	s.ServeHTTP(res, req)

	if res.Code != http.StatusNotFound {
		t.Error("Wrong response code. Expected 404 got", res.Code)
	}
}

func TestServeMux_HandleGet(t *testing.T) {
	s := NewServeMux()

	s.Handle("/a", rightHandler)
	req := httptest.NewRequest(http.MethodGet, "/a", nil)

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a" {
		t.Error("Wrong path, expected /a, got", path)
	}
}

func TestServeMux_HandlePost(t *testing.T) {
	s := NewServeMux()

	s.Handle("/a", rightHandler)
	req := httptest.NewRequest(http.MethodPost, "/a", nil)

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}

	if path != "/a" {
		t.Error("Wrong path, expected /a, got", path)
	}
}

func TestServeMux_MiddlewareSingle(t *testing.T) {
	s := NewServeMux()

	s.Middleware("/", mid1)
	s.Handle("/", rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 1, got", len(mids))
	}

	if mids[0] != mid1 {
		t.Error("wat")
	}
}

func TestServeMux_MiddlewareExceptFor(t *testing.T) {
	s := NewServeMux()

	s.MiddlewareExceptFor("/", mid1, http.MethodOptions, http.MethodPatch)
	s.Handle("/", rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 1, got", len(mids))
	}

	req = httptest.NewRequest(http.MethodOptions, "/", nil)

	_, mids, _ = s.HandlerAndMiddleware(req)

	if len(mids) != 0 {
		t.Fatal("Wrong number of middlewares returned. Expected 0, got", len(mids))
	}

	req = httptest.NewRequest(http.MethodPatch, "/", nil)

	_, mids, _ = s.HandlerAndMiddleware(req)

	if len(mids) != 0 {
		t.Fatal("Wrong number of middlewares returned. Expected 0, got", len(mids))
	}
}

func TestServeMux_MiddlewareExceptForOptions(t *testing.T) {
	s := NewServeMux()

	s.Route("/").MiddlewareExceptForOptions(mid1)
	s.Handle("/", rightHandler)

	req := httptest.NewRequest(http.MethodOptions, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 0 {
		t.Fatal("Wrong number of middlewares returned. Expected 0, got", len(mids))
	}
}

func TestServeMux_MiddlewareFor(t *testing.T) {
	s := NewServeMux()

	s.MiddlewareFor("/", mid1, http.MethodOptions, http.MethodPatch)
	s.Handle("/", rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 0 {
		t.Fatal("Wrong number of middlewares returned. Expected 0, got", len(mids))
	}

	req = httptest.NewRequest(http.MethodOptions, "/", nil)

	_, mids, _ = s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 1, got", len(mids))
	}

	req = httptest.NewRequest(http.MethodPatch, "/", nil)

	_, mids, _ = s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 1, got", len(mids))
	}
}

func TestServeMux_MiddlewareFor_Panic(t *testing.T) {
	s := NewServeMux()

	defer func() {
		err := recover()
		if err == nil {
			t.Error("Didn't panic")
			return
		}
	}()

	s.MiddlewareFor("/", mid1, http.MethodOptions, "llama")
}

func TestServeMux_MiddlewareFor_Nop(t *testing.T) {
	s := NewServeMux()

	s.MiddlewareFor("/", mid1)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 0 {
		t.Fatal("Wrong number of middlewares returned. Expected 0 got", len(mids))
	}
}

func TestServeMux_MiddlewareExceptFor_Any(t *testing.T) {
	s := NewServeMux()

	s.MiddlewareExceptFor("/", mid1)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 1 got", len(mids))
	}
}

func TestServeMux_MiddlewareExceptFor_None(t *testing.T) {
	s := NewServeMux()

	s.MiddlewareExceptFor("/", mid1,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodConnect,
		http.MethodPut,
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 0 {
		t.Fatal("Wrong number of middlewares returned. Expected 0 got", len(mids))
	}
}
func TestServeMux_MiddlewareDouble(t *testing.T) {
	s := NewServeMux()

	s.Route("/").
		Middleware(mid1).
		Get(rightHandler).
		Middleware(mid2)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 2 {
		t.Fatal("Wrong number of middlewares returned. Expected 2, got", len(mids))
	}

	if mids[0] != mid1 {
		t.Error("Wrong middleware 1")
	}
	if mids[1] != mid2 {
		t.Error("Wrong middleware 2")
	}
}

func TestServeMux_MiddlewareFunc(t *testing.T) {
	s := NewServeMux()

	var called bool

	midFunc := func(res http.ResponseWriter, req *http.Request, next func(http.ResponseWriter, *http.Request)) {
		called = true
	}

	s.MiddlewareFunc("/", midFunc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 2, got", len(mids))
	}

	s.ServeHTTP(nil, req)

	if !called {
		t.Error("Middleware not called")
	}
}

func TestServeMux_RequestPath(t *testing.T) {
	s := NewServeMux()

	var reqPath string

	s.Route("/users/:id/info").GetFunc(func(res http.ResponseWriter, req *http.Request) {
		reqPath = RequestPath(req)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/andrew/info", nil)

	s.ServeHTTP(nil, req)

	if reqPath != "/users/:id/info" {
		t.Error("Wrong request path returned", reqPath)
	}
}

func TestPathParams(t *testing.T) {
	s := NewServeMux()

	var params map[string]string

	handler := func(res http.ResponseWriter, r *http.Request) {
		params = PathParams(r)
	}

	s.Route("/:a/:b/:c/").GetFunc(handler)

	req := httptest.NewRequest(http.MethodGet, "/andrew/a/burian", nil)

	s.ServeHTTP(nil, req)

	if len(params) != 3 {
		t.Error("Wrong number of params returned", len(params))
	}

	if params["a"] != "andrew" {
		t.Error("Wrong value for a,", params["a"])
	}

	if params["b"] != "a" {
		t.Error("Wrong value for b,", params["b"])
	}

	if params["c"] != "burian" {
		t.Error("Wrong value for c,", params["c"])
	}
}

func TestPathParamsImmutable(t *testing.T) {
	s := NewServeMux()

	handler := func(res http.ResponseWriter, r *http.Request) {
		params := PathParams(r)
		params["id"] = "wrong"
		params = PathParams(r)
		if params["id"] == "wrong" {
			t.Error("Path params aren't immutable")
		}
	}

	s.Route("/:id/").GetFunc(handler)

	req := httptest.NewRequest(http.MethodGet, "/andrew", nil)

	s.ServeHTTP(nil, req)
}

func containsStr(strs []string, s string) int {
	for i, str := range strs {
		if strings.HasPrefix(str, s+"\t") {
			return i
		}
	}
	return -1
}
func TestServeMux_String(t *testing.T) {
	s := NewServeMux()

	s.Route("/").NotFound(rightHandler)

	s.Route("/empty")

	s.Route("/depth/one").Any(rightHandler)

	s.Route("/multi").Get(rightHandler).Post(rightHandler)

	s.RouteHost("example.com", "/another").Any(rightHandler)

	str := strings.Trim(s.String(), "\n")
	routes := strings.Split(str, "\n")

	t.Logf("String:\n%s", str)

	// right number of statements
	if len(routes) != 4 {
		t.Error("Wrong number of routes returned", len(routes))
	}

	// omit empty routes
	if containsStr(routes, "/empty") != -1 {
		t.Error("Empty route returned")
	}

	// root handler included properly
	if containsStr(routes, "/") != 0 {
		t.Error("Root route not included", containsStr(routes, "/"))
	}

	// host route included properly
	if containsStr(routes, "example.com/another") == 0 {
		t.Error("Did not inlcude host specific route")
	}

}

func TestServeMux_NotFound(t *testing.T) {
	s := NewServeMux()
	s.NotFound(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/llama", nil)

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong not found handler returned")
	}

	// not found should return an empty pattern
	if path != "" {
		t.Error("Wrong path returned", path)
	}
}

func TestServeMux_NotFoundDepth(t *testing.T) {
	s := NewServeMux()
	s.NotFound(wrongHandler)
	s.Route("/get").NotFound(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/get/llama", nil)

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong not found handler returned")
	}

	// not found should return an empty pattern
	if path != "" {
		t.Error("Wrong path returned", path)
	}
}

func TestServeMux_RouteHost(t *testing.T) {
	s := NewServeMux()

	s.RouteHost("example.com", "/").Get(rightHandler)
	s.Route("/").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Host = "example.com"

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}
	if path != "/" {
		t.Error("Wrong path set")
	}
}

func TestServeMux_HandleHost(t *testing.T) {
	s := NewServeMux()

	s.HandleHost("example.com", "/", rightHandler)
	s.Handle("/", wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Host = "example.com"

	h, path := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}
	if path != "/" {
		t.Error("Wrong path set")
	}
}

func TestServeMux_MiddlewareHost(t *testing.T) {
	s := NewServeMux()

	s.MiddlewareHost("example.com", "/", mid1)
	s.HandleHost("example.com", "/", rightHandler)

	s.Middleware("/", mid2)
	s.Handle("/", wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Host = "example.com"

	_, mids, _ := s.HandlerAndMiddleware(req)

	if len(mids) != 1 {
		t.Fatal("Wrong number of middlewares returned. Expected 1, got", len(mids))
	}

	if mids[0] != mid1 {
		t.Error("wat")
	}
}

func TestServeMux_HandleFunc(t *testing.T) {
	s := NewServeMux()

	rightFunc := dummyHandlerFunc("handlefunc")
	s.HandleFunc("/", rightFunc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Body.String() != "handlefunc" {
		t.Error("Body doesn't match")
	}
}

func TestServeMux_ServeHTTPHost(t *testing.T) {
	s := NewServeMux()

	s.RouteHost("example.com", "/").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.URL.Host = "example.com"
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Body.String() != "right" {
		t.Error("Wrong handler executed")
	}
}
