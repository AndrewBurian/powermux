package main

import (
	"fmt"
	"github.com/andrewburian/powermux"
	"gopkg.in/pg.v5"
	"net/http"
)

// UserHandler is a powermux handler. Any objects that can be statically shared between requests are kept
// in the object itself, in this case the database pool.
//
// The handler doesn't implement the ServeHTTP interface, so instead individual functions need to be bound into
// Powermux to route the requests. The UserHander does expose a Setup function where it does all this binding
// to keep the logic in one place.
type UserHandler struct {
	db *pg.DB
}

// Sets up a user handler with all the required functions
//
// Note that this function takes a powermux Route, not the entire ServeMux. This is so that it can be agnostic
// about the path leading up to it, but still have complete control over it's section of the route tree.
func (h *UserHandler) Setup(r *powermux.Route) {

	// using path parameters
	// these functions don't know or care what the route is above them
	r.Route("/:id").GetFunc(h.Get)

	// use the root of this section of the route tree
	r.PostFunc(h.CreateUser)
}

func (h *UserHandler) Get(rw http.ResponseWriter, req *http.Request) {
	// use the static database connection from the handler
	fmt.Println(h.db.String())

	// get a path parameter
	fmt.Println(powermux.PathParam(req, "id"))

	// use the request-scoped log entry from the logging library middleware
	entry := getLogEntry(req)
	entry.WithField("function", "UserHandler.Get").Info("Did something!")
}

func (h *UserHandler) CreateUser(rw http.ResponseWriter, req *http.Request) {
	// create the user
	entry := getLogEntry(req)
	entry.Info("User Created")
}
