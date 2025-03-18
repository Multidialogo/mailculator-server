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

func TestHandleMailQueueInvalidMethod(t *testing.T) {
	// Create a new HTTP request with an invalid method (GET instead of POST)
	req, err := http.NewRequest(http.MethodGet, "/email-queues", nil)
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
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rr.Code)
	}
}

func TestHandleMailQueueMissingData(t *testing.T) {
	// Create a request with missing email data
	requestPayload := []byte(`{"data": []}`)
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

	// Assert the response code is 400 Bad Request (missing email data)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleMailQueueInvalidJSON(t *testing.T) {
	// Create an invalid JSON request (malformed JSON)
	requestPayload := []byte(`{data: []}`)
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

	// Assert the response code is 400 Bad Request (invalid JSON)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleMailQueueInvalidIDFormat(t *testing.T) {
	// Create a request with an invalid ID format (not "userID:queueUUID:messageUUID")
	requestPayload := []byte(`
	{
		"data": [
			{
				"id": "invalid-id-format",
				"type": "email",
				"from": "test@example.com",
				"to": "recipient@example.com",
				"subject": "Test Subject"
			}
		]
	}`)
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

	// Assert the response code is 400 Bad Request (invalid ID format)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleMailQueueInvalidType(t *testing.T) {
	// Create a request with an invalid type (not "email")
	requestPayload := []byte(`
	{
		"data": [
			{
				"id": "user1:queue1:message1",
				"type": "invalid-type",
				"from": "test@example.com",
				"to": "recipient@example.com",
				"subject": "Test Subject"
			}
		]
	}`)
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

	// Assert the response code is 400 Bad Request (invalid type)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}

func TestHandleMailQueueMissingRequiredFields(t *testing.T) {
	// Create a request with missing required fields (e.g., missing "from", "to", or "subject")
	requestPayload := []byte(`
	{
		"data": [
			{
				"id": "user1:queue1:message1",
				"type": "email",
				"from": "",
				"to": "recipient@example.com",
				"subject": "Test Subject"
			}
		]
	}`)
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

	// Assert the response code is 400 Bad Request (missing "from" field)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rr.Code)
	}
}
