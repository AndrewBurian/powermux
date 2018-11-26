package main

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/twinj/uuid"
	"net/http"
)

// LoggerMiddleware is designed to act like a powermux compatible library
//
// It exposes an object that can be instantiated with any static properties needed to operate,
// then the object can be registered as a middleware to inject data into the request context.
//
// It then exposes helper functions used for interacting with requests that have passed through its middleware.
type LoggerMiddleware struct {
	// the parent event common to all requests
	baseEntry *logrus.Entry
}

// Middleware libraries should always use their own context key type to prevent context key collisions between
// different middlewares
type logCtxKeyType string

var logCtxKey = logCtxKeyType("event")

// Injects a new log entry with a request UUID into the request context
func (m *LoggerMiddleware) ServeHTTPMiddleware(rw http.ResponseWriter, req *http.Request, next func(rw http.ResponseWriter, req *http.Request)) {

	// inject the log into the context along with some info
	entry := m.baseEntry.WithField("id", uuid.NewV4())

	req = req.WithContext(context.WithValue(req.Context(), logCtxKey, entry))

	next(rw, req)
}

// Gets the data out of the request context for use
func getLogEntry(req *http.Request) *logrus.Entry {
	return req.Context().Value(logCtxKey).(*logrus.Entry)
}
