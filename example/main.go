package main

import (
	"github.com/AndrewBurian/powermux"
	"github.com/sirupsen/logrus"
	"gopkg.in/pg.v5"
	"net/http"
)

func main() {

	// create the router
	mux := powermux.NewServeMux()

	// setup the logging helper
	logger := &LoggerMiddleware{
		baseEntry: logrus.NewEntry(logrus.StandardLogger()).WithField("project", "Powermux-sample"),
	}

	// add the logging middleware
	mux.Route("/").Middleware(logger)

	// set up static resources like the database
	dbConnection := pg.Connect(&pg.Options{})

	// create the static route handlers like this users handler
	usersHandler := UserHandler{
		db: dbConnection,
	}

	// Get the sub route
	usersRoute := mux.Route("/users")

	// let the handler add its own routes to the mux under the prefix we give it
	usersHandler.Setup(usersRoute)

	// create a handler object for static content
	staticHandler := http.FileServer(http.Dir("/var/www/static"))

	// register the whole object instead of bound functions
	// as the fileServer implements the http.Handler interface
	mux.Route("/static").Any(staticHandler)

	// Serve
	http.ListenAndServe(":8080", mux)
}
