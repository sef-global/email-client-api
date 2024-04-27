package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"sync"
	"time"

	"github.com/go-mail/mail/v2"
	"github.com/mayura-andrew/email-client/internal/data"
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

type EmailData struct {
	Subject   string
	Body      string
	Recipient string
}

// func New(host string, port int, username)

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dailer: dialer,
		sender: sender,
	}
}

// func New(host string, port int, username)

func NewMail(e data.EmailModel, host string, port int, username, password, sender, subject string, recipients []string, body string) (map[string]*EmailStatus, error) {

	d := mail.NewDialer(host, port, username, password)

	emailStatuses := make(map[string]*EmailStatus)

	var statusMutex sync.Mutex

	// create a channel to queue the recipients
	queue := make(chan string)
	// create an WaitGroup to wait for all emails to be sent
	var wg sync.WaitGroup

	email := &data.Email{
		Sender:  sender,
		Body:    body,
		Subject: subject,
	}

	// start a number of worker goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for recipient := range queue {

				tmpl, err := template.ParseFiles("email_template.tmpl")

				if err != nil {
					log.Println(err)
					return
				}

				data := EmailData{
					Subject:   subject,
					Body:      body,
					Recipient: recipient,
				}

				bodyBuf := new(bytes.Buffer)
				err = tmpl.ExecuteTemplate(bodyBuf, "htmlBody", data)

				if err != nil {
					log.Println(err)
					return
				}

				m := mail.NewMessage()
				m.SetHeader("From", sender)
				m.SetHeader("To", recipient)
				m.SetHeader("Subject", subject)

				// modifiedBody := body + "\n<img src=\"http://localhost:4000/api/v1/track?email=" + recipient + "\" width=\"1\" height=\"1\" />"

				m.SetBody("text/html", bodyBuf.String()) // Join the elements of the body slice into a single string

				err = d.DialAndSend(m)
				if err != nil {
					fmt.Println("Failed to send test email to -> " + recipient + ": " + err.Error())
				} else {
					fmt.Println("Sent test email successfully to -> " + recipient)
					statusMutex.Lock()
					emailStatuses[recipient].Sent = true
					statusMutex.Unlock()
					err := e.InsertEmail(email, recipient, emailStatuses[recipient].Sent, emailStatuses[recipient].SentTime)
					if err != nil {
						log.Println(err)
						return
					}
				}
			}
		}()
	}

	// Enqueue the recipients and increment the WaitGroup counter
	for _, recipient := range recipients {
		queue <- recipient

		// Initialize the status of the email
		statusMutex.Lock()
		emailStatuses[recipient] = &EmailStatus{
			Sent:     false,
			Opened:   false,
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

	return emailStatuses, nil
}

func UpdateEmailTracking(e data.EmailModel, recipient string) error {
	return e.UpdateEmail(recipient)
}
