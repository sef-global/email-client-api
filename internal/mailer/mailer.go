package mailer

import (
	// "encoding/json"
	"bytes"
	"embed"
	"html/template"
	"time"
	// "net/http"
	"github.com/go-mail/mail/v2"
)


var templateFS embed.FS

type Mailer struct {
	dailer  *mail.Dialer
	sender string
}

// func New(host string, port int, username)

func New(host string, port int, username, password, sender string) Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dailer:  dialer,
		sender:  sender,
	}
}


func (m Mailer) Send(recipients, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/" + templateFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmplBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmplBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipients)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetHeader("text/plain", plainBody.String())
	msg.SetHeader("text/html", htmplBody.String())

	err = m.dailer.DialAndSend(msg)
	if err != nil {
		return err
	}
	return nil

	
}
// func (app *application) sendEmailHandler(w http.ResponseWriter, r *http.Request) {
//     // Parse the request body
//     var req struct {
//         Sender     string   `json:"sender"`
//         Recipients []string `json:"recipients"`
// 		Subject string `json:"subject"`
//         Body       []string   `json:"body"`
//     }

// 		err := json.NewDecoder(r.Body).Decode(&req)
// 		if err != nil {
// 			http.Error(w, "Invalid request body", http.StatusBadRequest)
// 			return
// 		}

// 		// Call the New function
// 		_, err = New(app.config.smtp.host, app.config.smtp.port, app.config.smtp.username, app.config.smtp.password, req.Sender, req.Subject, req.Recipients, req.Body)
// 		if err != nil {
// 			http.Error(w, "Failed to send email", http.StatusInternalServerError)
// 			return
// 		}

//     // Send the response
//     w.WriteHeader(http.StatusOK)
//     w.Write([]byte("Email sent successfully."))

