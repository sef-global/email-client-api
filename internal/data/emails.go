package data

import (
	"context"
	"database/sql"
	"errors"
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

type EmailRecipient struct {
	ID int64 `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Sender string `json:"sender"`
	Recipient string `json:"recipients"`
	Body string `json:"body"`  
	Subject string `json:"subject"`
	Status string `json:"status"`
	SentTime time.Time `json:"sentTime"`
	Opened string `json:"opened"`
	OpenedTime CustomNullTime `json:"openedTime"`
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

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()


	return e.DB.QueryRowContext(ctx, query, args...).Scan(&email.ID, &recipient, &isSent, &sentTime, false)
}

func (e EmailModel) Get(recipient string) (*EmailRecipient, error) {
	if len(recipient) < 1 {
		return nil, ErrRecordNotFound
	} 
	query := `SELECT recipients.id, recipients.recipient, recipients.status, recipients.sent_time, recipients.opened, recipients.opened_time, emails.created_at, emails.sender, emails.body, emails.subject FROM recipients JOIN emails ON recipients.email_id = emails.id 
	WHERE recipients.recipient = $1;`

	var email EmailRecipient

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()
	
	err := e.DB.QueryRowContext(ctx, query, recipient).Scan(&email.ID, &email.Recipient, &email.Status, &email.SentTime, &email.Opened, &email.OpenedTime, &email.CreatedAt, &email.Sender, &email.Body, &email.Subject)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
		return nil, err
		}
	}
	return &email, nil
}

func (e EmailModel) UpdateEmail(recipient string) error {
	query := `UPDATE recipients SET opened = true, opened_time = $1 WHERE recipient = $2`

	args := []any{time.Now(), recipient}


	_, err := e.DB.Exec(query, args...)
	return err
}

func (e EmailModel) Delete(id int64) error {
	return nil
}
