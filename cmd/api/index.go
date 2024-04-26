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
	<!DOCTYPE html>
	<html>
	<head>
		<title>Welcome to Our API</title>
		<style>
			body {
				font-family: Arial, sans-serif;
				margin: 0;
				padding: 0;
				background-color: #f0f0f0;
			}
			.container {
				width: 80%;
				margin: auto;
				padding: 20px;
			}
			h1, h2 {
				color: #333;
			}
			p {
				color: #666;
			}
		</style>
	</head>
	<body>
		<div class="container">
			<h1>Welcome to Our API!</h1>
			<h2>API Endpoints</h2>
			<p><strong>GET /v1/healthcheck:</strong> Check the health of the application.</p>
			<h2>Contact</h2>
			<p>Visit our website: <a href="https://mayuraandrew.tech">https://mayuraandrew.tech</a></p>
		</div>
	</body>
	</html>
    `)
}
