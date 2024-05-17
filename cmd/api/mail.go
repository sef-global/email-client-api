package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/mayura-andrew/email-client/internal/data"
	"github.com/mayura-andrew/email-client/internal/mailer"
	"github.com/mayura-andrew/email-client/internal/validator"
)

type Mailer struct {
	dailer *mail.Dialer
	sender string
}

type EmailStatus struct {
	Sent     bool
	Opened   bool
	SentTime time.Time
}

func (app *application) sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	// Parse the request body
	var req struct {
		Sender     string   `json:"sender"`
		Recipients []string `json:"recipients"`
		Subject    string   `json:"subject"`
		Body       string   `json:"body"`
	}

	err := app.readJSON(w, r, &req)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	email := &data.Email{
		Sender:     req.Sender,
		Recipients: req.Recipients,
		Subject:    req.Subject,
		Body:       req.Body,
	}

	v := validator.New()

	if data.ValidateEmail(v, email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	emailStatus, err := mailer.NewMail(app.models.Emails, app.config.smtp.host, app.config.smtp.port, app.config.smtp.username, app.config.smtp.password, app.config.smtp.sender, req.Subject, req.Recipients, req.Body)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/emails/%d", email.ID))

	err = app.writeJSON(w, http.StatusCreated, envelop{"status": emailStatus}, headers)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}

func (app *application) track(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	fmt.Println(id)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelop{"status": map[string]string{"error": "Missing id parameter"}}, nil)
		return
	}
	
	log.Printf("Email opened: %d", id)
	
	err = mailer.UpdateEmailTracking(app.models.Emails, id)
	if err != nil {
		log.Printf("Failed to update email tracking: %v", err)
		app.writeJSON(w, http.StatusBadRequest, envelop{"status": map[string]string{"error": "Internal server error"}}, nil)
	}
	redirectURL := "https://scholarx.sefglobal.org"
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func (app *application) showEmailHandler(w http.ResponseWriter, r *http.Request) {

	recipient := path.Base(r.URL.Path)
	if recipient == "" {
		app.notFoundResponse(w, r)
		return
	} 

	details, err := app.models.Emails.GetAllSent()
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorRespone(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"emails": details}, nil)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}
}
