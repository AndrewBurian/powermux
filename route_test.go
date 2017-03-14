package powermux

import "testing"

func TestRoute_RouteIsSame(t *testing.T) {
	r := &Route{
		pattern: "/",
	}

	r2 := r.Route("/")

	if r2 == nil {
		t.Fatal("Route returend nil pointer")
	}

	if r != r2 {
		t.Error("Route did not return self pointer")
	}
}

func TestRoute_RouteAdd(t *testing.T) {
	r := &Route{
		pattern: "/",
	}

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

	if r2.isRoot {
		t.Error("Route thinks it's a rooted tree")
	}
}

func TestRoute_RouteAddRetrieve(t *testing.T) {
	r := &Route{
		pattern: "/",
	}

	r1 := r.Route("/a")
	r2 := r.Route("/a")

	if r1 != r2 {
		t.Fatal("Did not return existing reference")
	}
}
