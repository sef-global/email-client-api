package data

import (
	"database/sql"
	"encoding/json"
	"time"
)

type CustomNullTime struct {
	sql.NullTime
}

func (nt *CustomNullTime) MarshalJSON() ([]byte, error) {
	if nt.Valid {
		return json.Marshal(struct {
			Time  time.Time `json:"time"`
			Valid bool      `json:"valid"`
		}{
			Time:  nt.Time,
			Valid: nt.Valid,
		})
	}
	return json.Marshal(nil)
}

func (nt *CustomNullTime) UnmarshalJSON(b []byte) error {
	var aux struct {
		Time  time.Time `json:"time"`
		Valid bool      `json:"valid"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	nt.Time = aux.Time
	nt.Valid = aux.Valid
	return nil
}
