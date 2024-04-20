package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/mayura-andrew/email-client/internal/data"
	"github.com/mayura-andrew/email-client/internal/validator"
)


type Mailer struct {
	dailer  *mail.Dialer
	sender string
}

type EmailStatus struct {
	Sent bool
	Opened bool
	SentTime time.Time
}


// func New(host string, port int, username)

func New(host string, port int, username, password, sender, subject string, recipients []string, body string) (*Mailer, error) {

	d := mail.NewDialer(host, port, username, password)

	emailStatuses := make(map[string]*EmailStatus)

	var statusMutex sync.Mutex

	// create a channel to queue the recipients
	queue := make(chan string)
	// create an WaitGroup to wait for all emails to be sent
	var wg sync.WaitGroup




	// start a number of worker goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for recipient := range queue {
				m := mail.NewMessage()
				m.SetHeader("From", sender)
				m.SetHeader("To", recipient)
				m.SetHeader("Subject", subject)
			
				body += "\n<img src=\"http://localhost:4000/api/v1/track?email=" + recipient + "\" width=\"1\" height=\"1\" />"
			
				m.SetBody("text/html", body) // Join the elements of the body slice into a single string
			
				err := d.DialAndSend(m)
				if err != nil {
                    fmt.Println("Failed to send test email to -> " + recipient + ": " + err.Error())
				} else {
					fmt.Println("Sent test email successfully to -> " + recipient)
					statusMutex.Lock()
					emailStatuses[recipient].Sent = true
					statusMutex.Unlock()
				}

				// Decrement the waitGroup counter
				wg.Done()
			}
		}()
	}

		// Enqueue the recipients and increment the WaitGroup counter
		for _, recipient := range recipients {
			wg.Add(1)
			queue <- recipient
	
			// Initialize the status of the email
			statusMutex.Lock()
			emailStatuses[recipient] = &EmailStatus{
				Sent: false,
				Opened: false,
				SentTime: time.Now(),
			}
	
			statusMutex.Unlock()
		}
	// Close the channel to signal that no more recipients will be enqueued 	
	close(queue)

	// Wait for all emails to be sent
	wg.Wait()

	for recipient, status := range emailStatuses {
        log.Printf("Email to %s: sent=%v, opened=%v, sentTime=%v", recipient, status.Sent, status.Opened, status.SentTime)
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

	err := app.readJSON(w, r, &req) 
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	email := &data.Email {
		Sender: req.Sender,
		Recipients: req.Recipients,
		Subject: req.Subject,
		Body: req.Body,
	}

	v := validator.New()
	
	if data.ValidateEmail(v, email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Emails.InsertEmail(email)
	if err != nil {
		app.serverErrorRespone(w, r, err)
		return
	}


	
	fmt.Fprintf(w, "%+v\n", req)

	// err := json.NewDecoder(r.Body).Decode(&req)
	// if err != nil {
	// 	http.Error(w, "Invalid request body", http.StatusBadRequest)
	// 	return
	// }

	// Call the New function
	_, err = New(app.config.smtp.host, app.config.smtp.port, app.config.smtp.username, app.config.smtp.password, req.Sender, req.Subject, req.Recipients, req.Body)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/api/v1/emails/%d", email.ID))

	err = app.writeJSON(w, http.StatusCreated, envelop{"email": email}, headers)
	if err != nil {
		app.serverErrorRespone(w, r, err)
	}

    // // Send the response
    // w.WriteHeader(http.StatusOK)
    // w.Write([]byte("Email sent successfully."))
}


// email tracking

func (app *application) track(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Missing email parameter", http.StatusBadRequest)
		return
	}
	log.Printf("Email opened: %s", email)
	w.Header().Set("Content-Type", "image/gif")
	w.Write([]byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\xff\xff\xff,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;"))
}