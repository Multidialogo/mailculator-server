package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"mailculator/internal/outbox"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"mailculator/internal/config"
	"mailculator/internal/testutils"
)

// Define the simplified structure of the JSON
type RequestData struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

var payloadsDir string
var expectationsDir string
var db *dynamodb.Client
var outboxServiceTest *outbox.Outbox

func init() {
	// Get the directory where the test source is located (i.e., the directory of this test file)
	_, currentFilePath, _, _ := runtime.Caller(0)
	rootDir := filepath.Dir(currentFilePath)
	testDir := filepath.Join(rootDir, "testData")
	payloadsDir = filepath.Join(testDir, "payloads")
	expectationsDir = filepath.Join(testDir, "expectations")

	awsConfig := aws.Config{
		Region: os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
		BaseEndpoint: aws.String(os.Getenv("AWS_BASE_ENDPOINT")),
	}
	db = dynamodb.NewFromConfig(awsConfig)
	outboxServiceTest = outbox.NewOutbox(db)
}

func TestHandleMailQueue(t *testing.T) {
	functionName := testutils.GetCleanFunctionName()
	testPayloadDir := filepath.Join(payloadsDir, functionName)

	// Load test payload
	requestPayload, err := os.ReadFile(filepath.Join(testPayloadDir, "request.json"))
	if err != nil {
		t.Fatalf("Failed to read request file: %v", err)
	}

	registry := config.GetRegistry()
	inputPath = filepath.Join(registry.Get("APP_DATA_PATH"), registry.Get("INPUT_PATH"))

	// Load fixture files (attachments)
	testutils.LoadFixturesFilesInInputDirectory(filepath.Join(testPayloadDir, "files"), filepath.Join(inputPath, functionName), t)

	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodPost, "/email-queues", bytes.NewBuffer(requestPayload))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/vnd.api+json")

	// Create a recorder to capture the HTTP response
	rr := httptest.NewRecorder()

	// Call the handler
	handler := http.HandlerFunc(handleMailQueue)
	handler.ServeHTTP(rr, req)

	// Assert the response code is 201 Created
	if rr.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", rr.Code)
	}

	responsePayload, err := os.ReadFile(filepath.Join(expectationsDir, functionName, "response.json"))
	if err != nil {
		t.Fatalf("Failed to read file response file: %v", err)
	}

	// Compare the actual and expected JSON
	assert.JSONEq(t, string(responsePayload), rr.Body.String())

	// Assert that a .eml file exists at in drafts dir
	draftOutputPath := filepath.Join(registry.Get("APP_DATA_PATH"), registry.Get("OUTBOX_PATH"))

	var requestData RequestData
	err = json.Unmarshal(requestPayload, &requestData)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	messagePath, err := testutils.GenerateMessagePath(requestData.Data[0].ID)
	if err != nil {
		t.Fatalf("Failed to generate message path: %v", err)
	}

	emlFilePath := filepath.Join(draftOutputPath, fmt.Sprintf("%s.EML", messagePath))
	assert.FileExists(t, emlFilePath, "Expected .eml file to exist at %s, but it does not.", emlFilePath)

	// Checking new entries in outbox table
	res, err := outboxServiceTest.Query(context.TODO(), outbox.StatusProcessing, 25)
	assert.NoError(t, err)
	assert.Len(t, res, 2)

	cleanUpDb(t)
}

func cleanUpDb(t *testing.T) {
	query := fmt.Sprintf("SELECT Id, Status FROM \"%v\"", "Outbox")
	stmt := &dynamodb.ExecuteStatementInput{Statement: aws.String(query)}
	res, err := db.ExecuteStatement(context.TODO(), stmt)
	assert.NoError(t, err)

	var items []outbox.EmailItemRow
	_ = attributevalue.UnmarshalListOfMaps(res.Items, &items)

	query = fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", "Outbox")
	for _, item := range items {
		params, _ := attributevalue.MarshalList([]interface{}{item.Id, item.Status})
		stmt = &dynamodb.ExecuteStatementInput{Statement: aws.String(query), Parameters: params}
		_, err = db.ExecuteStatement(context.TODO(), stmt)
		assert.NoError(t, err)
	}
}

func TestHandleMailQueueInvalidRequest(t *testing.T) {
	tests := []struct {
		name             string
		requestMethod    string
		requestBody      []byte
		expectedHttpCode int
		expectedHttpBody string
	}{
		{
			name:          "invalid request method",
			requestMethod: http.MethodGet,
			requestBody: []byte(`
{
	"data": [
        {
			"id": "65ed6bfa-063c-5219-844d-e099c88a17f4",
			"type": "email",
			"from": "user@example.com",
			"reply_to": "user@example.com",
			"to": "user@example.com",
			"subject": "string",
			"body_html": "string",
			"body_text": "string"
		}
	]
}`),
			expectedHttpCode: http.StatusMethodNotAllowed,
			expectedHttpBody: "Invalid request method\n",
		},
		{
			name:             "Missing request data",
			requestMethod:    http.MethodPost,
			requestBody:      []byte(`{"data": []}`),
			expectedHttpCode: http.StatusBadRequest,
			expectedHttpBody: "Invalid request body: Key: 'QueueCreationAPI.Data' Error:Field validation for 'Data' failed on the 'gt' tag\n",
		},
		{
			name:             "Invalid json",
			requestMethod:    http.MethodPost,
			requestBody:      []byte(`{data: []}`),
			expectedHttpCode: http.StatusBadRequest,
			expectedHttpBody: "Error unmarshalling request body: invalid character 'd' looking for beginning of object key string\n",
		},
		{
			name:          "Invalid id format",
			requestMethod: http.MethodPost,
			requestBody: []byte(`
{
	"data": [
        {
			"id": "invalid_id_format",
			"type": "email",
			"from": "user@example.com",
			"reply_to": "user@example.com",
			"to": "user@example.com",
			"subject": "string",
			"body_html": "string",
			"body_text": "string"
		}
	]
}`),
			expectedHttpCode: http.StatusBadRequest,
			expectedHttpBody: "Invalid request body: Key: 'QueueCreationAPI.Data[0].ID' Error:Field validation for 'ID' failed on the 'uuid' tag\n",
		},
		{
			name:          "Ivalid type",
			requestMethod: http.MethodPost,
			requestBody: []byte(`
{
	"data": [
        {
			"id": "65ed6bfa-063c-5219-844d-e099c88a17f4",
			"type": "invalid-type",
			"from": "user@example.com",
			"reply_to": "user@example.com",
			"to": "user@example.com",
			"subject": "string",
			"body_html": "string",
			"body_text": "string"
		}
	]
}`),
			expectedHttpCode: http.StatusBadRequest,
			expectedHttpBody: "Invalid request body: Key: 'QueueCreationAPI.Data[0].Type' Error:Field validation for 'Type' failed on the 'eq' tag\n",
		},
		{
			name:          "Ivalid recipient email",
			requestMethod: http.MethodPost,
			requestBody: []byte(`
{
	"data": [
        {
			"id": "65ed6bfa-063c-5219-844d-e099c88a17f4",
			"type": "email",
			"from": "user@example.com",
			"reply_to": "user@example.com",
			"to": "invalid_email_string",
			"subject": "string",
			"body_html": "string",
			"body_text": "string",
			"attachments": [],
			"custom_headers": {
				"property1": "string",
				"property2": "string"
			},
			"callback_on_success": "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"OK\"}' https://mycallback.it",
      		"callback_on_failure": "curl -X POST -H \"Content-Type: application/json\" -d '{\"status\": \"KO\"}' https://mycallback.it"
		}
	]
}`),
			expectedHttpCode: http.StatusBadRequest,
			expectedHttpBody: "Invalid request body: Key: 'QueueCreationAPI.Data[0].To' Error:Field validation for 'To' failed on the 'email' tag\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.requestMethod, "/email-queues", bytes.NewBuffer(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/vnd.api+json")

			// Create a recorder to capture the HTTP response
			rr := httptest.NewRecorder()

			// Call the handler
			handler := http.HandlerFunc(handleMailQueue)
			handler.ServeHTTP(rr, req)

			// Assert the response code is 405 Method Not Allowed
			if rr.Code != tt.expectedHttpCode {
				t.Errorf("Expected status %d, got %d", tt.expectedHttpCode, rr.Code)
			}

			if rr.Body.String() != tt.expectedHttpBody {
				t.Errorf("Expected body %s, got %s", tt.expectedHttpBody, rr.Body.String())
			}
		})
	}
}
