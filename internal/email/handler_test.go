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

func (m *emailServiceMock) GetStaleEmails(_ context.Context) ([]Email, error) {
	return nil, nil
}

func (m *emailServiceMock) GetInvalidEmails(_ context.Context) ([]Email, error) {
	return nil, nil
}

func (m *emailServiceMock) RequeueEmail(_ context.Context, _ string) error {
	return nil
}

func (m *emailServiceMock) ScanAndSetTTL(_ context.Context, _ int64, _ int) (*ScanAndSetTTLResult, error) {
	return &ScanAndSetTTLResult{ProcessedRecords: 100, TotalRecords: 100, HasMoreRecords: false}, nil
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

func TestScanAndSetTTLHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		requestBody      string
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:           "valid request - 200",
			requestBody:    `{"ttl_timestamp": 1735689600, "max_records": 1000}`,
			expectedStatus: http.StatusOK,
			expectedResponse: `{
				"processed_records": 100,
				"total_records": 100,
				"has_more_records": false
			}`,
		},
		{
			name:             "invalid json - 400",
			requestBody:      `{"ttl_timestamp": "invalid", "max_records": 1000}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error": "error unmarshalling request body: json: cannot unmarshal string into Go struct field scanAndSetTTLRequest.ttl_timestamp of type int64"}`,
		},
		{
			name:             "validation error - ttl_timestamp too small - 400",
			requestBody:      `{"ttl_timestamp": 0, "max_records": 1000}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error": "error validating request body: Key: 'scanAndSetTTLRequest.TTLTimestamp' Error:Field validation for 'TTLTimestamp' failed on the 'required' tag"}`,
		},
		{
			name:             "validation error - max_records too small - 400",
			requestBody:      `{"ttl_timestamp": 1735689600, "max_records": 0}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error": "error validating request body: Key: 'scanAndSetTTLRequest.MaxRecords' Error:Field validation for 'MaxRecords' failed on the 'required' tag"}`,
		},
		{
			name:             "validation error - max_records too large - 400",
			requestBody:      `{"ttl_timestamp": 1735689600, "max_records": 10001}`,
			expectedStatus:   http.StatusBadRequest,
			expectedResponse: `{"error": "error validating request body: Key: 'scanAndSetTTLRequest.MaxRecords' Error:Field validation for 'MaxRecords' failed on the 'max' tag"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/scan-and-set-ttl", bytes.NewReader([]byte(tc.requestBody)))
			response := httptest.NewRecorder()

			service := newEmailServiceMock(nil)
			sut := NewScanAndSetTTLHandler(service)

			sut.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatus, response.Code)
			assert.JSONEq(t, tc.expectedResponse, response.Body.String())
		})
	}
}
