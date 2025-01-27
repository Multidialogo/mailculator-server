package API

import (
	"testing"
)

func TestValidateRequest(t *testing.T) {
	tests := []struct {
		name          string
		apiRequest    *QueueCreationAPI
		expectedError string
	}{
		{
			name: "Valid request",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From          string            `json:"from"`
						ReplyTo       string            `json:"replyTo"`
						To            string            `json:"to"`
						Subject       string            `json:"subject"`
						BodyHTML      string            `json:"bodyHTML"`
						BodyText      string            `json:"bodyText"`
						Attachments   []string          `json:"attachments"`
						CustomHeaders map[string]string `json:"customHeaders"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          string            `json:"from"`
							ReplyTo       string            `json:"replyTo"`
							To            string            `json:"to"`
							Subject       string            `json:"subject"`
							BodyHTML      string            `json:"bodyHTML"`
							BodyText      string            `json:"bodyText"`
							Attachments   []string          `json:"attachments"`
							CustomHeaders map[string]string `json:"customHeaders"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
						},
					},
				},
			},
			expectedError: "",
		},
		{
			name: "Missing 'from' field",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From          string            `json:"from"`
						ReplyTo       string            `json:"replyTo"`
						To            string            `json:"to"`
						Subject       string            `json:"subject"`
						BodyHTML      string            `json:"bodyHTML"`
						BodyText      string            `json:"bodyText"`
						Attachments   []string          `json:"attachments"`
						CustomHeaders map[string]string `json:"customHeaders"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          string            `json:"from"`
							ReplyTo       string            `json:"replyTo"`
							To            string            `json:"to"`
							Subject       string            `json:"subject"`
							BodyHTML      string            `json:"bodyHTML"`
							BodyText      string            `json:"bodyText"`
							Attachments   []string          `json:"attachments"`
							CustomHeaders map[string]string `json:"customHeaders"`
						}{
							To:       "to@example.com",
							ReplyTo:  "reply@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
						},
					},
				},
			},
			expectedError: "missing 'from' field",
		},
		{
			name: "Invalid email in 'from' field",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From          string            `json:"from"`
						ReplyTo       string            `json:"replyTo"`
						To            string            `json:"to"`
						Subject       string            `json:"subject"`
						BodyHTML      string            `json:"bodyHTML"`
						BodyText      string            `json:"bodyText"`
						Attachments   []string          `json:"attachments"`
						CustomHeaders map[string]string `json:"customHeaders"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          string            `json:"from"`
							ReplyTo       string            `json:"replyTo"`
							To            string            `json:"to"`
							Subject       string            `json:"subject"`
							BodyHTML      string            `json:"bodyHTML"`
							BodyText      string            `json:"bodyText"`
							Attachments   []string          `json:"attachments"`
							CustomHeaders map[string]string `json:"customHeaders"`
						}{
							From:     "invalid-email",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
						},
					},
				},
			},
			expectedError: "invalid 'from' email address",
		},
		{
			name: "Invalid ID format",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From          string            `json:"from"`
						ReplyTo       string            `json:"replyTo"`
						To            string            `json:"to"`
						Subject       string            `json:"subject"`
						BodyHTML      string            `json:"bodyHTML"`
						BodyText      string            `json:"bodyText"`
						Attachments   []string          `json:"attachments"`
						CustomHeaders map[string]string `json:"customHeaders"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID",
						Type: "email",
						Attributes: struct {
							From          string            `json:"from"`
							ReplyTo       string            `json:"replyTo"`
							To            string            `json:"to"`
							Subject       string            `json:"subject"`
							BodyHTML      string            `json:"bodyHTML"`
							BodyText      string            `json:"bodyText"`
							Attachments   []string          `json:"attachments"`
							CustomHeaders map[string]string `json:"customHeaders"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
						},
					},
				},
			},
			expectedError: "invalid ID format, expected 'userID:queueUUID:messageUUID'",
		},
		{
			name: "Missing 'subject' field",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From          string            `json:"from"`
						ReplyTo       string            `json:"replyTo"`
						To            string            `json:"to"`
						Subject       string            `json:"subject"`
						BodyHTML      string            `json:"bodyHTML"`
						BodyText      string            `json:"bodyText"`
						Attachments   []string          `json:"attachments"`
						CustomHeaders map[string]string `json:"customHeaders"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          string            `json:"from"`
							ReplyTo       string            `json:"replyTo"`
							To            string            `json:"to"`
							Subject       string            `json:"subject"`
							BodyHTML      string            `json:"bodyHTML"`
							BodyText      string            `json:"bodyText"`
							Attachments   []string          `json:"attachments"`
							CustomHeaders map[string]string `json:"customHeaders"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
						},
					},
				},
			},
			expectedError: "missing 'subject' field",
		},
		{
			name: "Either bodyHTML or bodyText must be provided",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From          string            `json:"from"`
						ReplyTo       string            `json:"replyTo"`
						To            string            `json:"to"`
						Subject       string            `json:"subject"`
						BodyHTML      string            `json:"bodyHTML"`
						BodyText      string            `json:"bodyText"`
						Attachments   []string          `json:"attachments"`
						CustomHeaders map[string]string `json:"customHeaders"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          string            `json:"from"`
							ReplyTo       string            `json:"replyTo"`
							To            string            `json:"to"`
							Subject       string            `json:"subject"`
							BodyHTML      string            `json:"bodyHTML"`
							BodyText      string            `json:"bodyText"`
							Attachments   []string          `json:"attachments"`
							CustomHeaders map[string]string `json:"customHeaders"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "",
							BodyText: "",
						},
					},
				},
			},
			expectedError: "either 'bodyHTML' or 'bodyText' must be provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRequest(tt.apiRequest)
			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}
