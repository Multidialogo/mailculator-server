package email

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type requeueEmailServiceMock struct {
	returnErr error
	calledId  string
}

func newRequeueEmailServiceMock(withError bool) *requeueEmailServiceMock {
	if withError {
		return &requeueEmailServiceMock{returnErr: errors.New("mock error")}
	}
	return new(requeueEmailServiceMock)
}

func (m *requeueEmailServiceMock) RequeueEmail(_ context.Context, id string) error {
	m.calledId = id
	return m.returnErr
}

func TestRequeueEmailHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	type caseStruct struct {
		name               string
		withServiceError   bool
		emailId            string
		expectedStatusCode int
		expectedBody       string
	}

	testCases := []caseStruct{
		{
			name:               "success",
			withServiceError:   false,
			emailId:            "test-id-123",
			expectedStatusCode: http.StatusNoContent,
			expectedBody:       "",
		},
		{
			name:               "service error",
			withServiceError:   true,
			emailId:            "test-id-456",
			expectedStatusCode: http.StatusInternalServerError,
			expectedBody:       `{"error": "error requeuing email"}`,
		},
		{
			name:               "missing id",
			withServiceError:   false,
			emailId:            "",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error": "id parameter is required"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var request *http.Request
			if tc.emailId != "" {
				request = httptest.NewRequest(http.MethodPost, "/emails/"+tc.emailId+"/requeue", nil)
				request.SetPathValue("id", tc.emailId)
			} else {
				request = httptest.NewRequest(http.MethodPost, "/emails//requeue", nil)
			}
			
			response := httptest.NewRecorder()

			service := newRequeueEmailServiceMock(tc.withServiceError)
			sut := NewRequeueEmailHandler(service)

			sut.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatusCode, response.Code)
			
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, response.Body.String())
			}

			if tc.emailId != "" && !tc.withServiceError {
				assert.Equal(t, tc.emailId, service.calledId)
			}
		})
	}
}

