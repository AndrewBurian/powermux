package powermux

import "testing"

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