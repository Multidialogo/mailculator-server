package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"mailculator/internal/API"
	"mailculator/internal/config"
	"mailculator/internal/model"
	"mailculator/internal/service"
)

var emailQueueStorage *service.EmailQueueStorage

var inputPath string

// init function to initialize necessary services
func init() {
	registry := config.GetRegistry()
	basePath := registry.Get("APP_DATA_PATH")
	inputPath = filepath.Join(basePath, registry.Get("INPUT_PATH"))
	draftOutputPath := filepath.Join(basePath, registry.Get("DRAFT_OUTPUT_PATH"))
	outboxPath := filepath.Join(basePath, registry.Get("OUTBOX_PATH"))

	err := os.MkdirAll(draftOutputPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create email draft storage directory \"%s\": %v", draftOutputPath, err)
		os.Exit(1)
	}

	err = os.MkdirAll(inputPath, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create email draft storage directory \"%s\": %v", inputPath, err)
		os.Exit(1)
	}

	// Initialize the EmailQueueStorage service
	emailQueueStorage = service.NewEmailQueueStorage(outboxPath)
}

// main function to start the server and handle routes
func main() {
	http.HandleFunc("/email-queues", handleMailQueue)

	// Start the server
	port := ":8080"
	fmt.Printf("Starting server on %s\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}

// handleMailQueue handles incoming POST requests for mail-queues
func handleMailQueue(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Parse the JSON request body into the QueueCreationAPI struct
	var APIRequest API.QueueCreationAPI
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&APIRequest); err != nil {
		http.Error(w, fmt.Sprintf("Error unmarshalling request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := API.ValidateRequest(&APIRequest); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Process each email data in the request
	var emails []*model.Email
	var userID string
	var messageUUID string
	var queueUUID string

	for _, emailData := range APIRequest.Data {
		// Extract userID and messageUUID from the ID field
		ids := strings.Split(emailData.ID, ":")

		// Check if userID and messageUUID are already set and match the current ones
		if userID != "" && userID != ids[0] {
			http.Error(w, "User ID mismatch", http.StatusBadRequest)
			return
		}
		if queueUUID != "" && queueUUID != ids[1] {
			http.Error(w, "Queue UUID mismatch", http.StatusBadRequest)
			return
		}

		// Set the userID and messageUUID if they haven't been set yet
		if userID == "" {
			userID = ids[0]
		}
		if queueUUID == "" {
			queueUUID = ids[1]
		}
		if messageUUID == "" {
			messageUUID = ids[2]
		}

		attachmentPaths := make([]string, len(emailData.Attributes.Attachments))
		for i, attachmentPath := range emailData.Attributes.Attachments {
			attachmentPaths[i] = filepath.Join(inputPath, attachmentPath)
		}

		// Create a new email using the constructor
		email := model.NewEmail(
			userID,
			queueUUID,
			messageUUID,
			emailData.Attributes.From,
			emailData.Attributes.ReplyTo,
			emailData.Attributes.To,
			emailData.Attributes.Subject,
			emailData.Attributes.BodyHTML,
			emailData.Attributes.BodyText,
			attachmentPaths,
			emailData.Attributes.CustomHeaders,
			time.Now(),
		)

		// Append the created email to the slice
		emails = append(emails, email)
	}

	// Save the emails using the EmailQueueStorage service
	err := emailQueueStorage.SaveEmailsAsEML(emails)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving emails: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the 201 Created status and the location of the saved emails
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/vnd.api+json") // Set the correct content type

	// JSON:APIRequest-compliant response
	response := struct {
		Data struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		} `json:"data"`
	}{
		Data: struct {
			Type  string `json:"type"`
			ID    string `json:"id"`
			Links struct {
				Self string `json:"self"`
			} `json:"links"`
		}{
			Type: "mail-queue",
			ID:   fmt.Sprintf("%s:%s", userID, queueUUID),
			Links: struct {
				Self string `json:"self"`
			}{
				Self: fmt.Sprintf("/email-queues/%s:%s", userID, queueUUID),
			},
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
