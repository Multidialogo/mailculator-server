package main

import (
	"os"
	"bytes"
	"fmt"
	"path/filepath"
	"net/http"
	"net/http/httptest"
	"testing"
	"runtime"
	"encoding/json"

	"mailculator/internal/testutils"
	"mailculator/internal/config"
	"github.com/stretchr/testify/assert"
)

// Define the simplified structure of the JSON
type RequestData struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

var rootDir string
var payloadsDir string
var expectationsDir string

func init() {
	// Get the directory where the test source is located (i.e., the directory of this test file)
	_, currentFilePath, _, _ := runtime.Caller(0)
	rootDir = filepath.Dir(currentFilePath)
	testDir := filepath.Join(rootDir, "testData")
	payloadsDir = filepath.Join(testDir, "payloads")
	expectationsDir = filepath.Join(testDir, "expectations")
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
	req, err := http.NewRequest(http.MethodPost, "/email-queues", bytes.NewBuffer([]byte(requestPayload)))
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
	draftOutputPath := filepath.Join(registry.Get("APP_DATA_PATH"), registry.Get("DRAFT_OUTPUT_PATH"))

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
				"attributes": {
					"from": "test@example.com",
					"to": "recipient@example.com",
					"subject": "Test Subject"
				}
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
				"attributes": {
					"from": "test@example.com",
					"to": "recipient@example.com",
					"subject": "Test Subject"
				}
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
				"attributes": {
					"from": "",
					"to": "recipient@example.com",
					"subject": "Test Subject"
				}
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
