package entity

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"regexp"
	"time"
)

type Account struct {
	ID                   string     `json:"id"`
	Email                *string    `json:"email,omitempty"`
	CreatedAt            *time.Time `json:"created_at,omitempty"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`
	DisplayName          *string    `json:"display_name,omitempty"`
	WakeTime             *string    `json:"wake_time,omitempty"`  // 24-hour clock; hours and minutes (15:04)
	SleepTime            *string    `json:"sleep_time,omitempty"` // 24-hour clock; hours and minutes (15:04)
	NotificationInterval *int       `json:"notification_interval,omitempty"`
}

func (account Account) Validate() error {
	return validation.ValidateStruct(&account,
		validation.Field(&account.Email, is.Email),
		validation.Field(&account.WakeTime, validation.Match(regexp.MustCompile(`^(2[0-3]|[01]?[0-9]):([0-5]?[0-9])$`))),
		validation.Field(&account.SleepTime, validation.Match(regexp.MustCompile(`^(2[0-3]|[01]?[0-9]):([0-5]?[0-9])$`))),
		validation.Field(&account.NotificationInterval, validation.Min(0), validation.Max(24)),
	)
}
