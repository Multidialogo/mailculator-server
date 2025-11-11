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

type staleEmailsServiceMock struct {
	returnErr   error
	staleEmails []Email
}

func newStaleEmailsServiceMock(withError bool, staleEmails []Email) *staleEmailsServiceMock {
	if withError {
		return &staleEmailsServiceMock{returnErr: errors.New("mock error")}
	}
	return &staleEmailsServiceMock{staleEmails: staleEmails}
}

func (m *staleEmailsServiceMock) GetStaleEmails(_ context.Context) ([]Email, error) {
	if m.returnErr != nil {
		return nil, m.returnErr
	}
	return m.staleEmails, nil
}

func TestGetStaleEmailsHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	type caseStruct struct {
		name               string
		withServiceError   bool
		staleEmails        []Email
		expectedStatusCode int
		expectedBody       string
	}

	fixedTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	testCases := []caseStruct{
		{
			name:             "success with emails",
			withServiceError: false,
			staleEmails: []Email{
				{
					Id:        "test-id-1",
					Status:    "INTAKING",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime.Add(-1 * time.Hour),
				},
				{
					Id:        "test-id-2",
					Status:    "PROCESSING",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime.Add(-2 * time.Hour),
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[{"id":"test-id-1","status":"INTAKING","created_at":"2024-01-01T12:00:00Z","updated_at":"2024-01-01T11:00:00Z"},{"id":"test-id-2","status":"PROCESSING","created_at":"2024-01-01T12:00:00Z","updated_at":"2024-01-01T10:00:00Z"}]`,
		},
		{
			name:               "success with no emails",
			withServiceError:   false,
			staleEmails:        []Email{},
			expectedStatusCode: http.StatusOK,
			expectedBody:       `[]`,
		},
		{
			name:               "service error",
			withServiceError:   true,
			staleEmails:        nil,
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"error": "error getting stale emails"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/stale-emails", nil)
			response := httptest.NewRecorder()

			service := newStaleEmailsServiceMock(tc.withServiceError, tc.staleEmails)
			sut := NewGetStaleEmailsHandler(service)

			sut.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatusCode, response.Code)
			assert.JSONEq(t, tc.expectedBody, response.Body.String())
		})
	}
}
