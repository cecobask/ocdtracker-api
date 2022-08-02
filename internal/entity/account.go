package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"time"
)

type Account struct {
	ID                   string     `json:"id"`
	Email                *string    `json:"email,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
	DisplayName          *string    `json:"display_name,omitempty"`
	WakeTime             *time.Time `json:"wake_time,omitempty"`
	SleepTime            *time.Time `json:"sleep_time,omitempty"`
	NotificationInterval *int       `json:"notification_interval,omitempty"`
}

func (account Account) Validate() error {
	return validation.ValidateStruct(&account,
		validation.Field(&account.Email, is.Email),
		validation.Field(&account.NotificationInterval, validation.Min(0), validation.Max(24)),
	)
}
