package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"text/template"

	"github.com/go-mail/mail/v2"
)

var templateFS embed.FS

type Mailer struct {
	dailer  *mail.Dialer
	sender string
}

// func New(host string, port int, username)

func New(host string, port int, username, password, sender string, string subject, recipients []string, body string) (*Mailer, error) {
	d := mail.NewDialer(host, port, username, password)

	// Send a test email to each recipient to verify the SMTP server connection
	m := mail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", recipients...) // Set the "To" header to the slice of recipients
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	err := d.DialAndSend(m)
	if err != nil {
		return nil, fmt.Errorf("failed to send test email: %w", err)
	}

	fmt.Println("Sent test email successfully.")

	return &Mailer{
		dailer: d,
		sender: sender,
	}, nil
}

func (app *application) sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		Sender     string   `json:"sender"`
		Recipients []string `json:"recipients"`
		
		Body       string   `json:"body"`

	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Read the SMTP server details from the environment
	host := os.Getenv("SMTPHOST")
	port, _ := strconv.Atoi(os.Getenv("SMTPPORT")) // Convert the port to an integer
	username := os.Getenv("SMTPUSERNAME")
	password := os.Getenv("SMTPPASS")

	// Parse the JSON body into a map
	var bodyMap map[string]interface{}
	err = json.Unmarshal([]byte(req.Body), &bodyMap)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	// Load the template
	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
		return
	}

	// Execute the template with the bodyMap as the data
	var bodyBuffer bytes.Buffer
	err = tmpl.Execute(&bodyBuffer, bodyMap)
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}

	// Call the New function with the executed template as the body
	_, err = New(host, port, username, password, req.Sender, req.Recipients, bodyBuffer.String())
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	// Send the response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Email sent successfully."))
}