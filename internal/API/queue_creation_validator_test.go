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
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						From:                  "test@example.com",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "",
		},
		{
			name: "Missing 'from' field",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						To:                    "to@example.com",
						ReplyTo:               "reply@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "missing 'from' field",
		},
		{
			name: "Invalid email in 'from' field",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						From:                  "invalid-email",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "invalid 'from' email address",
		},
		{
			name: "Invalid ID format",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID",
						Type:                  "email",
						From:                  "test@example.com",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "invalid ID format, expected 'userID:queueUUID:messageUUID'",
		},
		{
			name: "Missing 'subject' field",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						From:                  "test@example.com",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "missing 'subject' field",
		},
		{
			name: "Either body_html or body_text must be provided",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						From:                  "test@example.com",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "",
						BodyText:              "",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "either 'body_html' or 'body_text' must be provided",
		},
		{
			name: "Invalid 'callback_on_success' curl command",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						From:                  "test@example.com",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "fake -X POST",
						CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
					},
				},
			},
			expectedError: "invalid 'callback_on_success' curl command",
		},
		{
			name: "Invalid 'callback_on_failure' curl command",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:                    "userID:queueUUID:messageUUID",
						Type:                  "email",
						From:                  "test@example.com",
						ReplyTo:               "reply@example.com",
						To:                    "to@example.com",
						Subject:               "Test Subject",
						BodyHTML:              "<p>Test HTML</p>",
						BodyText:              "Test Text",
						CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
						CallbackCallOnFailure: "fake -X POST",
					},
				},
			},
			expectedError: "invalid 'callback_on_failure' curl command",
		},
		{
			name: "Test callback_on_success and callback_on_failure attributes are optionals",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID                    string            `json:"id"`
					Type                  string            `json:"type"`
					From                  string            `json:"from"`
					ReplyTo               string            `json:"reply_to"`
					To                    string            `json:"to"`
					Subject               string            `json:"subject"`
					BodyHTML              string            `json:"body_html"`
					BodyText              string            `json:"body_text"`
					Attachments           []string          `json:"attachments"`
					CustomHeaders         map[string]string `json:"custom_headers"`
					CallbackCallOnSuccess string            `json:"callback_on_success"`
					CallbackCallOnFailure string            `json:"callback_on_failure"`
				}{
					{
						ID:       "userID:queueUUID:messageUUID",
						Type:     "email",
						From:     "test@example.com",
						ReplyTo:  "reply@example.com",
						To:       "to@example.com",
						Subject:  "Test Subject",
						BodyHTML: "<p>Test HTML</p>",
						BodyText: "Test Text",
					},
				},
			},
			expectedError: "",
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
