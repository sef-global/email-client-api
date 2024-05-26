package main

import (
	"fmt"
	"net/http"
)

func (app *application) rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, `
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
			<p><strong>GET /api/v1/healthcheck:</strong> Check the health of the application.</p>
			<p><strong>GET /debug/vars:</strong> Get debug variables.</p>
			<p><strong>GET /:</strong> Root endpoint.</p>
			<p><strong>POST /api/v1/send:</strong> Send an email.</p>
			<p><strong>GET /api/v1/track:</strong> Track an email.</p>
			<p><strong>GET /api/v1/recipients/:email:</strong> Get information about a specific email recipient.</p>
		</div>
	</body>
	</html>
    `)
}
