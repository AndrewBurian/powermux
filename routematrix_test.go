package powermux

import (
	"fmt"
	"testing"
	"net/http"
	"net/http/httptest"
	"bytes"
)

type RouteTest struct {
	route          string
	body           string
	method         string
	expectCode     int
	expectLocation string
}

var routeTests = []RouteTest{
	{
		route:      "/",
		expectCode: http.StatusNotFound,
	},
	{
		route:          "/foo/",
		expectLocation: "/foo",
	},
	{
		route:      "//example.com/",
		expectCode: http.StatusBadRequest,
	},
	{
		route:      "/foo//bar",
		expectCode: http.StatusBadRequest,
	},
}

func TestMuxRoutes(t *testing.T) {

	mux := NewServeMux()
	for _, tt := range routeTests {

		// sane defaults
		if tt.expectCode == 0 {
			if tt.expectLocation != "" {
				tt.expectCode = http.StatusPermanentRedirect
			} else {
				tt.expectCode = http.StatusOK
			}
		}
		if tt.method == "" {
			tt.method = http.MethodGet
		}

		t.Run(fmt.Sprintf("%s%s", tt.method, tt.route), func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.route, bytes.NewBufferString(tt.body))
			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectCode {
				t.Fatalf("Unexpected status code, expected=%d actual=%d", tt.expectCode, rec.Code)
			}

			if tt.expectLocation != "" && (rec.Code / 100) != 3 {
				t.Fatalf("Expected redirect, instead got code %d", rec.Code)
			}

			if tt.expectLocation != "" && tt.expectLocation != rec.Header().Get("Location") {
				t.Fatalf("Mismatched redirect. expected=%s, actual=%s", tt.expectLocation, rec.Header().Get("Location"))
			}
		})
	}
}
