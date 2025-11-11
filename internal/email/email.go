package email

import "time"

// Email represents an email record with its status and metadata
type Email struct {
	Id           string    `json:"id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ErrorMessage string    `json:"error_message,omitempty"`
}
