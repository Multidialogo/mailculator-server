package email

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"multicarrier-email-api/internal/eml"
)

type emailServiceMock struct {
	returnErr error
}

func newEmailServiceMock(withError bool) *emailServiceMock {
	if withError {
		return &emailServiceMock{returnErr: errors.New("mock error")}
	}
	return new(emailServiceMock)
}

func (m *emailServiceMock) Save(_ context.Context, _ []eml.EML) error {
	return m.returnErr
}

func TestCreateEmailHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	type caseStruct struct {
		name               string
		withServiceError   bool
		payloadFilePath    string
		expectedStatusCode int
		expectedBody       string
	}

	testCases := []caseStruct{
		{
			name:               "created",
			withServiceError:   false,
			payloadFilePath:    "testdata/handler_test/payloads/valid.json",
			expectedStatusCode: http.StatusCreated,
			expectedBody:       "{}",
		},
		{
			name:               "invalid json",
			withServiceError:   false,
			payloadFilePath:    "testdata/handler_test/payloads/wrong-property-types.json",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error": "error unmarshalling request body: json: cannot unmarshal number into Go struct field .data.id of type string"}`,
		},
		{
			name:               "validation errors",
			withServiceError:   false,
			payloadFilePath:    "testdata/handler_test/payloads/validation-errors.json",
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       `{"error": "error validating request body: Key: 'createEmailRequestBody.Data[0].BodyHTML' Error:Field validation for 'BodyHTML' failed on the 'required_without' tag\nKey: 'createEmailRequestBody.Data[0].BodyText' Error:Field validation for 'BodyText' failed on the 'required_without' tag"}`,
		},
		{
			name:               "service error",
			withServiceError:   true,
			payloadFilePath:    "testdata/handler_test/payloads/valid.json",
			expectedStatusCode: http.StatusConflict,
			expectedBody:       `{"error": "error saving emails"}`,
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

			service := newEmailServiceMock(tc.withServiceError)
			sut := NewCreateEmailHandler("testdata/handler_test/resources", service)

			sut.ServeHTTP(response, request)

			assert.Equal(t, tc.expectedStatusCode, response.Code)
			assert.JSONEq(t, tc.expectedBody, response.Body.String())
		})
	}
}
