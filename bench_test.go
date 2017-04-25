package powermux

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	MAX_WIDTH  = 500
	MAX_DEPTH  = 100
	FAN_DEPTH  = 4
	FAN_SPREAD = 8
)

func BenchmarkSingleRoute(b *testing.B) {
	r := NewServeMux()
	r.Route("/").Any(rightHandler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Handler(req)
	}
}

func BenchmarkShallowAndWide(b *testing.B) {
	r := NewServeMux()
	requests := make([]*http.Request, 0, MAX_WIDTH)
	for i := 0; i < MAX_WIDTH; i++ {
		route := "/" + hex.EncodeToString([]byte(fmt.Sprint(i)))
		r.Handle(route, rightHandler)
		requests = append(requests, httptest.NewRequest(http.MethodGet, route, nil))
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Handler(requests[i%MAX_WIDTH])
	}
}

// BenchmarkNarrowAndDeep is powermux's worst-case scenario. One route at the end
// of a really long path
func BenchmarkNarrowAndDeep(b *testing.B) {
	r := NewServeMux()
	var route string
	for i := 0; i < MAX_DEPTH; i++ {
		route += "/" + hex.EncodeToString([]byte(fmt.Sprint(i)))
	}
	r.Handle(route, rightHandler)
	req := httptest.NewRequest(http.MethodGet, route, nil)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Handler(req)
	}
}

func addFanRoutes(n int, r *Route) (routes []string) {
	for i := 0; i < FAN_SPREAD; i++ {
		route := "/" + hex.EncodeToString([]byte(fmt.Sprint(i)))
		r2 := r.Route(route)
		routes = append(routes, route)
		if n == 0 {
			break
		}
		addedRoutes := addFanRoutes(n-1, r2)
		routes = append(routes, addedRoutes...)
	}

	return routes
}

func BenchmarkFan(b *testing.B) {
	r := NewServeMux()
	routes := addFanRoutes(FAN_DEPTH, r.Route("/"))
	requests := make([]*http.Request, 0, len(routes))
	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		requests = append(requests, req)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		r.Handler(requests[i%len(requests)])
	}
}
