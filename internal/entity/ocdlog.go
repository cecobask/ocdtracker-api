package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"time"
)

type OCDLog struct {
	ID              uuid.UUID  `json:"id"`
	AccountID       string     `json:"account_id"`
	CreatedAt       *time.Time `json:"created_at,omitempty"`
	UpdatedAt       *time.Time `json:"updated_at,omitempty"`
	RuminateMinutes *int       `json:"ruminate_minutes,omitempty"`
	AnxietyLevel    *int       `json:"anxiety_level,omitempty"`
	Notes           *string    `json:"notes,omitempty"`
}

type OCDLogList struct {
	Logs       []OCDLog          `json:"logs"`
	Pagination PaginationDetails `json:"pagination"`
}

type PaginationDetails struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Count  int `json:"count"`
	Total  int `json:"total"`
}

func (ocdLog OCDLog) Validate() error {
	return validation.ValidateStruct(&ocdLog,
		validation.Field(&ocdLog.RuminateMinutes, validation.Min(0)),
		validation.Field(&ocdLog.AnxietyLevel, validation.Min(0), validation.Max(10)),
	)
}
