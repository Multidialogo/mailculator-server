package model

import (
	"reflect"
	"testing"
	"mailculator/internal/testutils"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		messageUUID   string
		to            string
		subject       string
		bodyHTML      string
		bodyText      string
		attachments   []string
		customHeaders map[string]string
		expectPanic   bool
		expectedEmail *Email
	}{
		{
			name:          "valid email creation",
			userID:        "user123",
			messageUUID:   "uuid123",
			to:            "recipient@example.com",
			subject:       "Subject",
			bodyHTML:      "<h1>HTML Body</h1>",
			bodyText:      "Plain Text Body",
			attachments:   []string{"file1.txt", "file2.txt"},
			customHeaders: map[string]string{"X-Custom-Header": "HeaderValue"},
			expectPanic:   false,
			expectedEmail: &Email{
				userID:        "user123",
				messageUUID:   "uuid123",
				to:            "recipient@example.com",
				subject:       "Subject",
				bodyHTML:      "<h1>HTML Body</h1>",
				bodyText:      "Plain Text Body",
				attachments:   []string{"file1.txt", "file2.txt"},
				customHeaders: map[string]string{"X-Custom-Header": "HeaderValue"},
				path:          "users/user123/messages/uuid123",
				date:          testutils.GetUnixEpoch(),
			},
		},
		{
			name:        "missing required fields (userID)",
			userID:      "",
			messageUUID: "uuid123",
			to:          "recipient@example.com",
			expectPanic: true,
		},
		{
			name:        "missing required fields (messageUUID)",
			userID:      "user123",
			messageUUID: "",
			to:          "recipient@example.com",
			expectPanic: true,
		},
		{
			name:        "missing required fields (to)",
			userID:      "user123",
			messageUUID: "uuid123",
			to:          "",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected panic, but did not panic")
					}
				}()
			}

			email := NewEmail(tt.userID, tt.messageUUID, tt.to, tt.subject, tt.bodyHTML, tt.bodyText, tt.attachments, tt.customHeaders, testutils.GetUnixEpoch())

			// Check if email matches expected values
			if !reflect.DeepEqual(email, tt.expectedEmail) {
				t.Errorf("expected %+v, but got %+v", tt.expectedEmail, email)
			}

			// Check immutability of attachments and customHeaders
			if !reflect.DeepEqual(email.Attachments(), tt.attachments) {
				t.Errorf("expected attachments %+v, but got %+v", tt.attachments, email.Attachments())
			}
			if !reflect.DeepEqual(email.CustomHeaders(), tt.customHeaders) {
				t.Errorf("expected customHeaders %+v, but got %+v", tt.customHeaders, email.CustomHeaders())
			}

			// Ensure that attachments and customHeaders are not modified
			tt.attachments[0] = "modified.txt"
			tt.customHeaders["X-Custom-Header"] = "modified"

			if reflect.DeepEqual(email.Attachments(), tt.attachments) {
				t.Errorf("attachments were modified when they should be immutable")
			}
			if reflect.DeepEqual(email.CustomHeaders(), tt.customHeaders) {
				t.Errorf("customHeaders were modified when they should be immutable")
			}
		})
	}
}

func TestEmailGetters(t *testing.T) {
	email := NewEmail("user123", "uuid123", "recipient@example.com", "Subject", "<h1>HTML</h1>", "Plain Text", []string{"file1.txt"}, map[string]string{"X-Custom-Header": "HeaderValue"}, testutils.GetUnixEpoch())

	tests := []struct {
		name   string
		value  interface{}
		getter func() interface{}
	}{
		{"UserID", "user123", func() interface{} { return email.UserID() }},
		{"MessageUUID", "uuid123", func() interface{} { return email.MessageUUID() }},
		{"To", "recipient@example.com", func() interface{} { return email.To() }},
		{"Subject", "Subject", func() interface{} { return email.Subject() }},
		{"BodyHTML", "<h1>HTML</h1>", func() interface{} { return email.BodyHTML() }},
		{"BodyText", "Plain Text", func() interface{} { return email.BodyText() }},
		{"Path", "users/user123/messages/uuid123", func() interface{} { return email.Path() }},
		{"Date", testutils.GetUnixEpoch(), func() interface{} { return email.Date() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.getter(); !reflect.DeepEqual(got, tt.value) {
				t.Errorf("expected %v, but got %v", tt.value, got)
			}
		})
	}
}
