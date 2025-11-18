package email

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type emailServiceMock struct {
	results []SaveResult
}

func newEmailServiceMock(results []SaveResult) *emailServiceMock {
	return &emailServiceMock{results: results}
}

func (m *emailServiceMock) Save(_ context.Context, requests []EmailRequest) []SaveResult {
	if m.results != nil {
		return m.results
	}

	// Default: all successful
	results := make([]SaveResult, len(requests))
	for i, req := range requests {
		results[i] = SaveResult{
			MessageId: req.MessageId,
			Success:   true,
		}
	}
	return results
}

func TestCreateEmailHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	type caseStruct struct {
		name               string
		serviceResults     []SaveResult
		payloadFilePath    string
		expectedStatusCode int
		expectedBody       string
	}

	testCases := []caseStruct{
		{
			name:               "all emails accepted - 201",
			serviceResults:     nil,
			payloadFilePath:    "testdata/handler_test/payloads/valid.json",
			expectedStatusCode: http.StatusCreated,
			expectedBody:       "{}",
		},
		{
			name: "at least one email accepted - 200",
			serviceResults: []SaveResult{
				{MessageId: "65ed6bfa-063c-5219-844d-e099c88a17f4", Success: true},
				{MessageId: "ff0fb587-e29b-4278-bbab-a525196b8917", Success: false, ErrorCode: ErrorCodeDuplicatedID, ErrorMessage: ErrorMessageDuplicatedID},
			},
			payloadFilePath:    "testdata/handler_test/payloads/valid.json",
			expectedStatusCode: http.StatusOK,
			expectedBody: `{
				"summary": {
					"total": 2,
					"successful": 1,
					"failed": 1
				},
				"results": [
					{
						"id": "65ed6bfa-063c-5219-844d-e099c88a17f4",
						"status": "success"
					},
					{
						"id": "ff0fb587-e29b-4278-bbab-a525196b8917",
						"status": "error",
						"error": {
							"code": "DUPLICATED_ID",
							"message": "Email with this ID already exists"
						}
					}
				]
			}`,
		},
		{
			name: "no emails accepted - 422",
			serviceResults: []SaveResult{
				{MessageId: "65ed6bfa-063c-5219-844d-e099c88a17f4", Success: false, ErrorCode: ErrorCodeDuplicatedID, ErrorMessage: ErrorMessageDuplicatedID},
				{MessageId: "ff0fb587-e29b-4278-bbab-a525196b8917", Success: false, ErrorCode: ErrorCodeDatabaseError, ErrorMessage: ErrorMessageDatabaseError},
			},
			payloadFilePath:    "testdata/handler_test/payloads/valid.json",
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedBody: `{
				"summary": {
					"total": 2,
					"successful": 0,
					"failed": 2
				},
				"results": [
					{
						"id": "65ed6bfa-063c-5219-844d-e099c88a17f4",
						"status": "error",
						"error": {
							"code": "DUPLICATED_ID",
							"message": "Email with this ID already exists"
						}
					},
					{
						"id": "ff0fb587-e29b-4278-bbab-a525196b8917",
						"status": "error",
						"error": {
							"code": "DATABASE_ERROR",
							"message": "Failed to save email to database"
						}
					}
				]
			}`,
		},
		{
			name:               "invalid json - 400",
			serviceResults:     nil,
			payloadFilePath:    "testdata/handler_test/payloads/wrong-property-types.json",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error": "error unmarshalling request body: json: cannot unmarshal number into Go struct field emailDataInput.data.id of type string"}`,
		},
		{
			name:               "validation errors - 400",
			serviceResults:     nil,
			payloadFilePath:    "testdata/handler_test/payloads/validation-errors.json",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error": "error validating request body: Key: 'createEmailRequestBody.Data[0].BodyHTML' Error:Field validation for 'BodyHTML' failed on the 'required_without' tag\nKey: 'createEmailRequestBody.Data[0].BodyText' Error:Field validation for 'BodyText' failed on the 'required_without' tag"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody, err := os.ReadFile(tc.payloadFilePath)
			if err != nil {
				t.Fatal(err)
			}

			request := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(requestBody))
			response := httptest.NewRecorder()

			service := newEmailServiceMock(tc.serviceResults)
			sut := NewCreateEmailHandler(service)

			sut.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatusCode, response.Code)
			assert.JSONEq(t, tc.expectedBody, response.Body.String())
		})
	}
}
