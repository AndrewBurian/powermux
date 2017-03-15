package powermux

import "net/http"

// Redirect adds a redirect handler for ANY method for this route
func (r *Route) Redirect(url string, permanent bool) *Route {
	var h http.Handler
	if permanent {
		h = http.RedirectHandler(url, http.StatusPermanentRedirect)
	} else {
		h = http.RedirectHandler(url, http.StatusTemporaryRedirect)
	}
	return r.Any(h)
}
