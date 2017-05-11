package powermux

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoute_RouteIsSame(t *testing.T) {
	r := newRoute()

	r2 := r.Route("/")

	if r2 == nil {
		t.Fatal("Route returend nil pointer")
	}

	if r != r2 {
		t.Error("Route did not return self pointer")
	}
}

func TestRoute_RouteAdd(t *testing.T) {
	r := newRoute()

	r2 := r.Route("/a")

	if r2 == nil {
		t.Fatal("Route returned nil pointer")
	}
	if r == r2 {
		t.Fatal("Route returned self pointer")
	}

	if r2.pattern != "a" {
		t.Error("Pattern not set")
	}

	if r2.isParam {
		t.Error("Route thinks it's a parameter")
	}

	if r2.isWildcard {
		t.Error("Route thinks it's a rooted tree")
	}
}

func TestRoute_RouteAddDepth(t *testing.T) {
	r := newRoute()

	r2 := r.Route("/a/b")
	r1 := r.Route("/a")

	if r2 == nil {
		t.Fatal("Route returned nil pointer")
	}
	if r == r2 {
		t.Fatal("Route returned self pointer")
	}

	if r2.pattern != "b" {
		t.Error("Pattern not set")
	}

	if r2.isParam {
		t.Error("Route thinks it's a parameter")
	}

	if r2.isWildcard {
		t.Error("Route thinks it's a rooted tree")
	}

	if r1.pattern != "a" {
		t.Error("Route got wrong pattern")
	}

	if len(r1.children) == 0 {
		t.Fatal("Route should have got children")
	}

	if r1.children[0] != r2 {
		t.Error("Wrong child assigned")
	}
}

func TestRoute_RouteAddRetrieve(t *testing.T) {
	r := newRoute()

	r1 := r.Route("/a")
	r2 := r.Route("/a")

	if r1 != r2 {
		t.Fatal("Did not return existing reference")
	}
}

func TestRoute_RouteAddRootedTree(t *testing.T) {
	r := newRoute()

	r1 := r.Route("/test")
	r2 := r.Route("/test/*")

	if r1 == r2 {
		t.Error("Routed tree didn't get it's own node")
	}

	if !r2.isWildcard {
		t.Error("Tree didn't get isWildcard flag")
	}

	if len(r2.children) != 0 {
		t.Error("Where did it get children?")
	}

	if r1.wildcardChild == nil {
		t.Fatal("R1 didn't get child")
	}

	if r1.wildcardChild != r2 {
		t.Error("Tree built incorrectly")
	}
}

func TestRoute_RouteDepth(t *testing.T) {
	r := newRoute()
	r2 := r.Route("/a/b")
	r1 := r.Route("/a")
	r3 := r1.Route("/b")

	if r2 != r3 {
		t.Error("'/a/b' != '/a'.'/b'")
		t.Log(r1)
	}
}

func TestRoute_TrailingSlash(t *testing.T) {
	r := newRoute()
	r1 := r.Route("/a/")

	if r1.pattern != "a" {
		t.Error("Pattern mismatch")
	}

	if r.children[0] != r1 {
		t.Error("Child misset")
	}

	if len(r.children[0].children) > 0 {
		t.Error("Unexpected grandchildren")
	}
}

func TestRoute_LeadingSlash(t *testing.T) {
	r := newRoute()
	r1 := r.Route("a")
	r2 := r.Route("/a")

	if r1 != r2 {
		t.Error("Route without leading slash not equivilant")
	}
}

func TestRoute_MethodHandlers(t *testing.T) {
	handlers := make(map[string]http.Handler)
	handlers[http.MethodGet] = dummyHandler("get")
	handlers[http.MethodPost] = dummyHandler("post")
	handlers[http.MethodConnect] = dummyHandler("connect")
	handlers[http.MethodHead] = dummyHandler("head")
	handlers[http.MethodDelete] = dummyHandler("delete")
	handlers[http.MethodOptions] = dummyHandler("options")
	handlers[http.MethodPatch] = dummyHandler("patch")
	handlers[http.MethodPut] = dummyHandler("put")

	s := NewServeMux()
	r := s.Route("/")
	r.Get(handlers[http.MethodGet])
	r.Post(handlers[http.MethodPost])
	r.Connect(handlers[http.MethodConnect])
	r.Head(handlers[http.MethodHead])
	r.Delete(handlers[http.MethodDelete])
	r.Options(handlers[http.MethodOptions])
	r.Patch(handlers[http.MethodPatch])
	r.Put(handlers[http.MethodPut])

	for method, handler := range handlers {
		req := httptest.NewRequest(method, "/", nil)
		h, _ := s.Handler(req)
		if h != handler {
			t.Error("Wrong handler for request type", method)
		}
	}
}

func TestRoute_MethodHandlerFuncs(t *testing.T) {
	handlers := make(map[string]http.HandlerFunc)
	handlers[http.MethodGet] = dummyHandlerFunc("get")
	handlers[http.MethodPost] = dummyHandlerFunc("post")
	handlers[http.MethodConnect] = dummyHandlerFunc("connect")
	handlers[http.MethodHead] = dummyHandlerFunc("head")
	handlers[http.MethodDelete] = dummyHandlerFunc("delete")
	handlers[http.MethodOptions] = dummyHandlerFunc("options")
	handlers[http.MethodPatch] = dummyHandlerFunc("patch")
	handlers[http.MethodPut] = dummyHandlerFunc("put")

	s := NewServeMux()
	r := s.Route("/")
	r.GetFunc(handlers[http.MethodGet])
	r.PostFunc(handlers[http.MethodPost])
	r.ConnectFunc(handlers[http.MethodConnect])
	r.HeadFunc(handlers[http.MethodHead])
	r.DeleteFunc(handlers[http.MethodDelete])
	r.OptionsFunc(handlers[http.MethodOptions])
	r.PatchFunc(handlers[http.MethodPatch])
	r.PutFunc(handlers[http.MethodPut])

	for method, h1 := range handlers {
		req := httptest.NewRequest(method, "/", nil)
		h2, _ := s.Handler(req)

		rec1 := httptest.NewRecorder()
		h1.ServeHTTP(rec1, req)

		rec2 := httptest.NewRecorder()
		h2.ServeHTTP(rec2, req)

		if rec2.Body.String() != rec1.Body.String() {
			t.Error("Handler response bodies don't match:", method,
				rec1.Body.String(), rec2.Body.String())
		}

		if rec2.Code != rec1.Code {
			t.Error("Handler response codes don't match:", method,
				rec1.Code, rec2.Code)
		}
	}
}

func TestRoute_AnyFunc(t *testing.T) {
	s := NewServeMux()
	r := s.Route("/")

	rightFunc := dummyHandlerFunc("any")

	r.AnyFunc(rightFunc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Body.String() != "any" {
		t.Error("Body doesn't match")
	}
}

func TestRoute_NotFoundFunc(t *testing.T) {
	s := NewServeMux()
	r := s.Route("/")

	rightFunc := dummyHandlerFunc("notFound")

	r.NotFoundFunc(rightFunc)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.ServeHTTP(rec, req)

	if rec.Body.String() != "notFound" {
		t.Error("Body doesn't match")
	}
}
