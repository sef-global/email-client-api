package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

type Models struct {
	Emails EmailModel
}

func NewModel(db *sql.DB) Models {
	return Models{
		Emails: EmailModel{DB: db},
	}
}
