package powermux

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type dummyHandler string

func (h dummyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, string(h))
}

var (
	rightHandler = dummyHandler("right")
	wrongHandler = dummyHandler("wrong")
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
		param = GetPathParam(req, "id")
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
		param = GetPathParam(req, "id")
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
		param = GetPathParam(req, "id")
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
		param = GetPathParam(req, "id")
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
