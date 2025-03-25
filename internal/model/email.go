package model

import (
	"fmt"
	"time"
)

// Email represents an immutable email with recipient, subject, body, attachments, and custom headers.
type Email struct {
	messageUUID           string
	from                  string
	replyTo               string
	to                    string
	subject               string
	bodyHTML              string
	bodyText              string
	attachments           []string
	customHeaders         map[string]string
	date                  time.Time
	path                  string
	callbackCallOnSuccess string
	callbackCallOnFailure string
}

// NewEmail creates a new immutable Email instance.
func NewEmail(messageUUID, from, replyTo, to, subject, bodyHTML, bodyText string, attachments []string, customHeaders map[string]string, date time.Time, callbackCallOnSuccess string, callbackCallOnFailure string) *Email {
	if messageUUID == "" || from == "" || replyTo == "" || to == "" {
		panic("messageUUID, from, replyTo, and to cannot be empty")
	}

	path := fmt.Sprintf("%d/%s/%s", date.Year(), date.Month(), messageUUID)

	// Clone slices and maps to enforce immutability
	clonedAttachments := make([]string, len(attachments))
	copy(clonedAttachments, attachments)

	clonedHeaders := make(map[string]string)
	for k, v := range customHeaders {
		clonedHeaders[k] = v
	}

	return &Email{
		messageUUID:           messageUUID,
		from:                  from,
		replyTo:               replyTo,
		to:                    to,
		subject:               subject,
		bodyHTML:              bodyHTML,
		bodyText:              bodyText,
		attachments:           clonedAttachments,
		customHeaders:         clonedHeaders,
		date:                  date,
		path:                  path,
		callbackCallOnSuccess: callbackCallOnSuccess,
		callbackCallOnFailure: callbackCallOnFailure,
	}
}

// MessageUUID returns the message UUID.
func (e *Email) MessageUUID() string { return e.messageUUID }

// From returns the sender's address.
func (e *Email) From() string { return e.from }

// ReplyTo returns the reply-to address.
func (e *Email) ReplyTo() string { return e.replyTo }

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

func (e *Email) CallbackCallOnSuccess() string { return e.callbackCallOnSuccess }

func (e *Email) CallbackCallOnFailure() string { return e.callbackCallOnFailure }
