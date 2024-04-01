package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)
	router.HandlerFunc(http.MethodGet, "/api/v1/healthcheck", app.healthcheckHandler)

	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	router.HandlerFunc(http.MethodGet, "/", app.rootHandler)

	router.HandlerFunc(http.MethodPost, "/api/v1/send-email", app.sendEmailHandler)

	return router

}