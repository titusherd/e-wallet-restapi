package entity

import "time"

type User struct {
	ID                  int        `json:"id"`
	Username            string     `json:"username"`
	Email               string     `json:"email"`
	PasswordHash        string     `json:"-"`
	ResetPasswordCode   *string    `json:"-"`
	ResetPasswordExpiry *time.Time `json:"-"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}
