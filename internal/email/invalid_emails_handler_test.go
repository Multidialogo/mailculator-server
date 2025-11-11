package email

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type invalidEmailsServiceMock struct {
	returnErr     error
	invalidEmails []Email
}

func newInvalidEmailsServiceMock(withError bool, invalidEmails []Email) *invalidEmailsServiceMock {
	if withError {
		return &invalidEmailsServiceMock{returnErr: errors.New("mock error")}
	}
	return &invalidEmailsServiceMock{invalidEmails: invalidEmails}
}

func (m *invalidEmailsServiceMock) GetInvalidEmails(_ context.Context) ([]Email, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.invalidEmails, nil
}

func TestGetInvalidEmailsHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	type caseStruct struct {
		name               string
		withServiceError   bool
		invalidEmails      []Email
		expectedStatusCode int
		expectedBody       string
	}

	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	testCases := []caseStruct{
		{
			name:             "success with emails",
			withServiceError: false,
			invalidEmails: []Email{
				{
					Id:           "test-id-1",
					Status:       "INVALID",
					CreatedAt:    fixedTime,
					UpdatedAt:    fixedTime.Add(-1 * time.Hour),
					ErrorMessage: "Invalid email address",
				},
				{
					Id:           "test-id-2",
					Status:       "INVALID",
					CreatedAt:    fixedTime,
					UpdatedAt:    fixedTime.Add(-2 * time.Hour),
					ErrorMessage: "Validation failed",
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"id":"test-id-1","status":"INVALID","created_at":"2024-01-01T12:00:00Z","updated_at":"2024-01-01T11:00:00Z","error_message":"Invalid email address"},{"id":"test-id-2","status":"INVALID","created_at":"2024-01-01T12:00:00Z","updated_at":"2024-01-01T10:00:00Z","error_message":"Validation failed"}]`,
		},
		{
			name:             "success with emails without error message",
			withServiceError: false,
			invalidEmails: []Email{
				{
					Id:        "test-id-3",
					Status:    "INVALID",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"id":"test-id-3","status":"INVALID","created_at":"2024-01-01T12:00:00Z","updated_at":"2024-01-01T12:00:00Z"}]`,
		},
		{
			name:               "success with no emails",
			withServiceError:   false,
			invalidEmails:      []Email{},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[]`,
		},
		{
			name:               "service error",
			withServiceError:   true,
			invalidEmails:      nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"error": "error getting invalid emails"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/invalid-emails", nil)
			response := httptest.NewRecorder()

			service := newInvalidEmailsServiceMock(tc.withServiceError, tc.invalidEmails)
			sut := NewGetInvalidEmailsHandler(service)

			sut.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatusCode, response.Code)
			assert.JSONEq(t, tc.expectedBody, response.Body.String())
		})
	}
}
