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
						From          			string            `json:"from"`
						ReplyTo       			string            `json:"replyTo"`
						To            			string            `json:"to"`
						Subject       			string            `json:"subject"`
						BodyHTML      			string            `json:"bodyHTML"`
						BodyText      			string            `json:"bodyText"`
						Attachments   			[]string          `json:"attachments"`
						CustomHeaders 			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
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
						From          			string            `json:"from"`
						ReplyTo       			string            `json:"replyTo"`
						To            			string            `json:"to"`
						Subject       			string            `json:"subject"`
						BodyHTML      			string            `json:"bodyHTML"`
						BodyText      			string            `json:"bodyText"`
						Attachments   			[]string          `json:"attachments"`
						CustomHeaders 			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							To:       "to@example.com",
							ReplyTo:  "reply@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
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
						From          			string            `json:"from"`
						ReplyTo       			string            `json:"replyTo"`
						To            			string            `json:"to"`
						Subject       			string            `json:"subject"`
						BodyHTML      			string            `json:"bodyHTML"`
						BodyText      			string            `json:"bodyText"`
						Attachments   			[]string          `json:"attachments"`
						CustomHeaders 			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "invalid-email",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
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
						From          			string            `json:"from"`
						ReplyTo       			string            `json:"replyTo"`
						To            			string            `json:"to"`
						Subject       			string            `json:"subject"`
						BodyHTML      			string            `json:"bodyHTML"`
						BodyText      			string            `json:"bodyText"`
						Attachments   			[]string          `json:"attachments"`
						CustomHeaders 			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
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
						From          			string            `json:"from"`
						ReplyTo       			string            `json:"replyTo"`
						To            			string            `json:"to"`
						Subject       			string            `json:"subject"`
						BodyHTML      			string            `json:"bodyHTML"`
						BodyText      			string            `json:"bodyText"`
						Attachments   			[]string          `json:"attachments"`
						CustomHeaders 			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
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
						From         			string            `json:"from"`
						ReplyTo      			string            `json:"replyTo"`
						To           			string            `json:"to"`
						Subject      			string            `json:"subject"`
						BodyHTML     			string            `json:"bodyHTML"`
						BodyText     			string            `json:"bodyText"`
						Attachments  			[]string          `json:"attachments"`
						CustomHeaders			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "",
							BodyText: "",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
						},
					},
				},
			},
			expectedError: "either 'bodyHTML' or 'bodyText' must be provided",
		},
		{
			name: "Invalid 'callbackCallOnSuccess' curl command",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From         			string            `json:"from"`
						ReplyTo      			string            `json:"replyTo"`
						To           			string            `json:"to"`
						Subject      			string            `json:"subject"`
						BodyHTML     			string            `json:"bodyHTML"`
						BodyText     			string            `json:"bodyText"`
						Attachments  			[]string          `json:"attachments"`
						CustomHeaders			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "fake -X POST",
							CallbackCallOnFailure: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it/",
						},
					},
				},
			},
			expectedError: "invalid 'callbackCallOnSuccess' curl command",
		},
		{
			name: "Invalid 'callbackCallOnFailure' curl command",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From         			string            `json:"from"`
						ReplyTo      			string            `json:"replyTo"`
						To           			string            `json:"to"`
						Subject      			string            `json:"subject"`
						BodyHTML     			string            `json:"bodyHTML"`
						BodyText     			string            `json:"bodyText"`
						Attachments  			[]string          `json:"attachments"`
						CustomHeaders			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
						}{
							From:     "test@example.com",
							ReplyTo:  "reply@example.com",
							To:       "to@example.com",
							Subject:  "Test Subject",
							BodyHTML: "<p>Test HTML</p>",
							BodyText: "Test Text",
							CallbackCallOnSuccess: "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it/",
							CallbackCallOnFailure: "fake -X POST",
						},
					},
				},
			},
			expectedError: "invalid 'callbackCallOnFailure' curl command",
		},
		{
			name: "Test callbackCallOnSuccess and callbackCallOnFailure attributes are optionals",
			apiRequest: &QueueCreationAPI{
				Data: []struct {
					ID         string `json:"id"`
					Type       string `json:"type"`
					Attributes struct {
						From         			string            `json:"from"`
						ReplyTo      			string            `json:"replyTo"`
						To           			string            `json:"to"`
						Subject      			string            `json:"subject"`
						BodyHTML     			string            `json:"bodyHTML"`
						BodyText     			string            `json:"bodyText"`
						Attachments  			[]string          `json:"attachments"`
						CustomHeaders			map[string]string `json:"customHeaders"`
						CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
						CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
					} `json:"attributes"`
				}{
					{
						ID:   "userID:queueUUID:messageUUID",
						Type: "email",
						Attributes: struct {
							From          			string            `json:"from"`
							ReplyTo       			string            `json:"replyTo"`
							To            			string            `json:"to"`
							Subject       			string            `json:"subject"`
							BodyHTML      			string            `json:"bodyHTML"`
							BodyText      			string            `json:"bodyText"`
							Attachments   			[]string          `json:"attachments"`
							CustomHeaders 			map[string]string `json:"customHeaders"`
							CallbackCallOnSuccess   string            `json:"callbackCallOnSuccess"`
							CallbackCallOnFailure   string            `json:"callbackCallOnFailure"`
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
