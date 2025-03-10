package model

import (
	"fmt"
	"time"
)

// Email represents an immutable email with recipient, subject, body, attachments, and custom headers.
type Email struct {
	userID                string
	queueUUID             string
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
func NewEmail(userID, queueUUID, messageUUID, from, replyTo, to, subject, bodyHTML, bodyText string, attachments []string, customHeaders map[string]string, date time.Time, callbackCallOnSuccess string, callbackCallOnFailure string) *Email {
	if userID == "" || queueUUID == "" || messageUUID == "" || from == "" || replyTo == "" || to == "" {
		panic("userID, queueUUID, messageUUID, from, replyTo, and to cannot be empty")
	}
	path := fmt.Sprintf("users/%s/queues/%s/messages/%s", userID, queueUUID, messageUUID)

	// Clone slices and maps to enforce immutability
	clonedAttachments := make([]string, len(attachments))
	copy(clonedAttachments, attachments)

	clonedHeaders := make(map[string]string)
	for k, v := range customHeaders {
		clonedHeaders[k] = v
	}

	return &Email{
		userID:                userID,
		queueUUID:             queueUUID,
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

// UserID returns the user ID.
func (e *Email) UserID() string { return e.userID }

// QueueUUID returns the queue UUID.
func (e *Email) QueueUUID() string { return e.queueUUID }

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
