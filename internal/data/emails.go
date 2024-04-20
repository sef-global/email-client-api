package data

import (
	"database/sql"
	"log"
	"time"

	"github.com/mayura-andrew/email-client/internal/validator"
)

type Email struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"-"`
	Sender string `json:"sender"`
	Recipients []string `json:"recipients,omitempty"`
	Body string `json:"body"`  
	Subject string `json:"Subject"`
	
}

type EmailModel struct {
	DB *sql.DB
}

func ValidateEmail(v *validator.Validator, email *Email) {
	v.Check(email.Sender != "", "sender", "must be provided")
	v.Check(len(email.Sender) >= 1, "sender", "must be more than 1 bytes long")
	v.Check(len(email.Recipients) != 0, "recipients", "must be provided")
	v.Check(len(email.Recipients) >= 1, "recipients", "must contain more than 1 recipient emails")
	v.Check(validator.Unique(email.Recipients), "recipients", "must not contain duplicate recipient emails")
	v.Check(email.Subject != "", "subject", "must be provided")
	v.Check(len(email.Subject) >= 1, "sender", "must be more than 1 bytes long")
	v.Check(email.Body != "", "body", "must be provided")
	v.Check(len(email.Body) >= 1, "body", "must be more than 1 bytes long")
}

func (e EmailModel) InsertEmail(email *Email, recipient string, isSent bool, sentTime time.Time) error {
	query := `INSERT INTO emails (sender, body, subject) VALUES ($1, $2, $3) RETURNING id, created_at`

	args := []any{email.Sender, email.Body, email.Subject}

	err := e.DB.QueryRow(query, args...).Scan(&email.ID, &email.CreatedAt)

	if err != nil {
		return err
	}
	err = e.InsertEmailRecipient(email, recipient, isSent, sentTime)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (e EmailModel) InsertEmailRecipient(email *Email, recipient string, isSent bool, sentTime time.Time) error {
	query := `INSERT INTO recipients (email_id, recipient, status, sent_time, opened)
	VALUES ($1, $2, $3, $4, $5)`

	args := []any{email.ID, recipient, isSent, sentTime, false}

	_, err := e.DB.Exec(query, args...)
	return err
}

func (e EmailModel) Get(id int64) (*Email, error) {
	return nil, nil
}

func (e EmailModel) Update(email *Email) error {
	return nil
}

func (e EmailModel) Delete(id int64) error {
	return nil
}