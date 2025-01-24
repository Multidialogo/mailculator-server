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

	emlFilePath := filepath.Join(draftOutputPath, fmt.Sprintf("%s.EML", requestData.Data[0].ID))
	assert.FileExists(t, emlFilePath, "Expected .eml file to exist at %s, but it does not.", emlFilePath)
}
