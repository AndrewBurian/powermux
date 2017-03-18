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

func TestServeMux_String(t *testing.T) {
	s := NewServeMux()
	s.Route("/a/b/c").Get(nil)
	s.Route("/d/e/f").Get(nil).Post(nil)
	s.Route("/a/b").Delete(nil).NotFound(nil)

	//str := s.String()
	//t.Error("\n"+str)
}

// Ensures that parameter routes have lower precedence than absolute routes
func TestServeMux_ParamPrecedence(t *testing.T) {
	s := NewServeMux()

	rightHandler := dummyHandler("right")
	wrongHandler := dummyHandler("wrong")

	s.Route("/users/:id/info").Get(wrongHandler)
	s.Route("/users/jim/info").Get(rightHandler)
	s.Route("/users/:id/detail").Get(wrongHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/jim/info", nil)
	h, _ := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}
}

// Ensures that wildcards have the lowest of all precedences
func TestServeMux_WildcardPrecedence(t *testing.T) {
	s := NewServeMux()

	rightHandler := dummyHandler("right")
	wrongHandler := dummyHandler("wrong")

	s.Route("/users/*").Get(wrongHandler)
	s.Route("/users/john").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/john", nil)
	h, _ := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}
}

// Ensures the wildcard handler isn't called when a path param was available
func TestServeMux_WildcardPathPrecedence(t *testing.T) {
	s := NewServeMux()

	rightHandler := dummyHandler("right")
	wrongHandler := dummyHandler("wrong")

	s.Route("/users/*").Get(wrongHandler)
	s.Route("/users/:id").Get(rightHandler)

	req := httptest.NewRequest(http.MethodGet, "/users/john", nil)
	h, _ := s.Handler(req)

	if h != rightHandler {
		t.Error("Wrong handler returned")
	}
}