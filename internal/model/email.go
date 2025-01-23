package model

import (
	"fmt"
	"time"
)

// Email represents an immutable email with recipient, subject, body, attachments, and custom headers.
type Email struct {
	userID        string
	messageUUID   string
	to            string
	subject       string
	bodyHTML      string
	bodyText      string
	attachments   []string
	customHeaders map[string]string
	date          time.Time
	path          string
}

// NewEmail creates a new immutable Email instance.
func NewEmail(userID, messageUUID, to, subject, bodyHTML, bodyText string, attachments []string, customHeaders map[string]string, date time.Time) *Email {
	if userID == "" || messageUUID == "" || to == "" {
		panic("userID, messageUUID, and to cannot be empty")
	}
	path := fmt.Sprintf("users/%s/messages/%s", userID, messageUUID)

	// Clone slices and maps to enforce immutability
	clonedAttachments := make([]string, len(attachments))
	copy(clonedAttachments, attachments)

	clonedHeaders := make(map[string]string)
	for k, v := range customHeaders {
		clonedHeaders[k] = v
	}

	return &Email{
		userID:        userID,
		messageUUID:   messageUUID,
		to:            to,
		subject:       subject,
		bodyHTML:      bodyHTML,
		bodyText:      bodyText,
		attachments:   clonedAttachments,
		customHeaders: clonedHeaders,
		date:          date,
		path:          path,
	}
}

// UserID returns the user ID.
func (e *Email) UserID() string { return e.userID }

// MessageUUID returns the message UUID.
func (e *Email) MessageUUID() string { return e.messageUUID }

// To returns the recipient address.
func (e *Email) To() string { return e.to }

// Subject returns the email subject.
func (e *Email) Subject() string { return e.subject }

// BodyHTML returns the HTML body.
func (e *Email) BodyHTML() string { return e.bodyHTML }

// BodyText returns the plain text body.
func (e *Email) BodyText() string { return e.bodyText }

// Attachments returns a copy of the attachments slice.
func (e *Email) Attachments() []string {
	cloned := make([]string, len(e.attachments))
	copy(cloned, e.attachments)
	return cloned
}

// CustomHeaders returns a copy of the custom headers map.
func (e *Email) CustomHeaders() map[string]string {
	cloned := make(map[string]string)
	for k, v := range e.customHeaders {
		cloned[k] = v
	}
	return cloned
}

// Date returns the email's date.
func (e *Email) Date() time.Time { return e.date }

// Path returns the calculated file path.
func (e *Email) Path() string { return e.path }
