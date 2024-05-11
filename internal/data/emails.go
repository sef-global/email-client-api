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
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"-"`
	Sender     string    `json:"sender"`
	Recipients []string  `json:"recipients,omitempty"`
	Body       string    `json:"body"`
	Subject    string    `json:"Subject"`
}

type EmailRecipient struct {
	ID         int64          `json:"id"`
	CreatedAt  time.Time      `json:"createdAt"`
	Sender     string         `json:"sender"`
	Recipient  string         `json:"recipients"`
	Body       string         `json:"body"`
	Subject    string         `json:"subject"`
	Status     string         `json:"status"`
	SentTime   time.Time      `json:"sentTime"`
	Opened     string         `json:"opened"`
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
	// v.Check(validator.Unique(email.Recipients), "recipients", "must not contain duplicate recipient emails")
	v.Check(email.Subject != "", "subject", "must be provided")
	v.Check(len(email.Subject) >= 1, "sender", "must be more than 1 bytes long")
	v.Check(email.Body != "", "body", "must be provided")
	v.Check(len(email.Body) >= 1, "body", "must be more than 1 bytes long")
}

func (e EmailModel) InsertEmail(email *Email, recipient string) (int64, error) {
	query := `INSERT INTO emails (sender, body, subject) VALUES ($1, $2, $3) RETURNING id, created_at`

	args := []any{email.Sender, email.Body, email.Subject}

	err := e.DB.QueryRow(query, args...).Scan(&email.ID, &email.CreatedAt)

	if err != nil {
		return 0, err
	}
	emailID, err := e.InsertEmailRecipient(email, recipient)
	if err != nil {
		log.Println(err)
	}

	return emailID, nil
}

func (e EmailModel) InsertEmailRecipient(email *Email, recipient string) (int64, error) {

	query := `INSERT INTO recipients (email_id, recipient, status, sent_time, opened)
	VALUES ($1, $2, $3, $4, $5) RETURNING id`

	args := []any{email.ID, recipient, false, time.Now(), false}

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()
	var id int64
	err := e.DB.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (e EmailModel) GetAllSent()(*[]EmailRecipient, error) {

	query := `SELECT recipients.id, recipients.recipient, recipients.status, recipients.sent_time, recipients.opened, recipients.opened_time, emails.created_at, emails.sender, emails.body, emails.subject FROM recipients JOIN emails ON recipients.email_id = emails.id;`

	ctx, cancle := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancle()

	rows, err := e.DB.QueryContext(ctx, query)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	defer rows.Close()

	var details []EmailRecipient
	
	for rows.Next() {
		d := EmailRecipient{}
	    err = rows.Scan(&d.ID, &d.Recipient, &d.Status, &d.SentTime, &d.Opened, &d.OpenedTime, &d.CreatedAt, &d.Sender, &d.Body, &d.Subject)
		if err != nil {
			return nil, err
		}
		details = append(details, d)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return &details, nil
}

func (e EmailModel) UpdateEmail(id int64) error {
	query := `UPDATE recipients SET opened = true, opened_time = $1 WHERE id = $2`

	args := []any{time.Now(), id}

	_, err := e.DB.Exec(query, args...)
	return err
}

func (e EmailModel) UpdateEmailStatus(id int64) error {

	query := `UPDATE recipients SET status = true , sent_time = $1 WHERE id = $2`

	args := []any{time.Now(), id}

	_, err := e.DB.Exec(query, args...)
	return err
}
