package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-mail/mail/v2"
)


type Mailer struct {
	dailer  *mail.Dialer
	sender string
}

// func New(host string, port int, username)

func New(host string, port int, username, password, sender, subject string, recipients []string, body string) (*Mailer, error) {
	d := mail.NewDialer(host, port, username, password)

	// Send a test email to each recipient to verify the SMTP server connection

	for _, recipient := range recipients {
		m := mail.NewMessage()
		m.SetHeader("From", sender)
		m.SetHeader("To", recipients...) // Set the "To" header to the slice of recipients
		m.SetHeader("Subject", subject)
	
		body += "\n<img src=\"http://localhost:4000/api/v1/track?email=" + recipient + "\" width=\"1\" height=\"1\" />"
	
		m.SetBody("text/html", body) // Join the elements of the body slice into a single string
	
		err := d.DialAndSend(m)
		if err != nil {
			return nil, fmt.Errorf("failed to send test email: %w", err)
		}
	
		fmt.Println("Sent test email successfully to -> " + recipient)
	}
	return &Mailer{
		dailer:     d,
		sender:     sender,
	}, nil
}


func (app *application) sendEmailHandler(w http.ResponseWriter, r *http.Request) {
    // Parse the request body
    var req struct {
        Sender     string   `json:"sender"`
        Recipients []string `json:"recipients"`
		Subject string `json:"subject"`
        Body       string   `json:"body"`
    }

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}


	if err != nil {
		fmt.Print(err)
		http.Error(w, "Faild to parse email template", http.StatusInternalServerError)
		return
	}






		// Call the New function
	_, err = New(app.config.smtp.host, app.config.smtp.port, app.config.smtp.username, app.config.smtp.password, req.Sender, req.Subject, req.Recipients, req.Body)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

    // Send the response
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Email sent successfully."))
}

func (app *application) emailSendHandler(w http.ResponseWriter, r *http.Request) {
	email := "mayuraalahakoon@gmail.com"
	data := "Test Email Content"
	err := app.mailer.Send(email, "test_email.tmpl",  data)
	if err != nil {
		app.serverErrorRespone(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelop{"data" : data}, nil)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}

// email tracking

func (app *application) track(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email parameter", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
    w.Write([]byte("User is opened our email:)"))
}