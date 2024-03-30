package main

import (
	"fmt"
	"net/http"
)

// !Important --> this heathcheckHandler is implemented as a method on application struct
func (app *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Go Lang API Project</title>
	</head>
	<body>
		<h2>API Endpoints</h2>
		<h3>Health Check</h3>
		<p>GET /v1/healthcheck: Check the health of the application.</p>
		<p>Visit my website := https://mayuraandrew.tech</p>
	</body>
	</html>
    `)

}
