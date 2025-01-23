package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"mailculator/internal/model"
	"mailculator/internal/service"
)

type EmailAPI struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			To            string            `json:"to"`
			Subject       string            `json:"subject"`
			BodyHTML      string            `json:"bodyHTML"`
			BodyText      string            `json:"bodyText"`
			Attachments   []string          `json:"attachments"`
			CustomHeaders map[string]string `json:"customHeaders"`
		} `json:"attributes"`
	} `json:"data"`
}

var emailQueueStorage *service.EmailQueueStorage

// init function to initialize necessary services
func init() {
	// Set up the base path for storing emails as EML files
	basePath := "./data/email_queues"
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		log.Fatalf("Failed to create email storage directory: %v", err)
	}

	// Initialize the EmailQueueStorage service
	emailQueueStorage = service.NewEmailQueueStorage(basePath)
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

	// Parse the JSON request body into the EmailAPI struct
	var emailAPI EmailAPI
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&emailAPI); err != nil {
		http.Error(w, fmt.Sprintf("Error unmarshalling request body: %v", err), http.StatusBadRequest)
		return
	}

	if len(emailAPI.Data) == 0 {
		http.Error(w, "No email data provided", http.StatusBadRequest)
		return
	}

	// Process each email data in the request
	var emails []*model.Email
	var userID string
	var messageUUID string

	for _, emailData := range emailAPI.Data {
		// Check if the type is "email"
		if emailData.Type != "email" {
			http.Error(w, fmt.Sprintf("Invalid type '%s', expected 'email'", emailData.Type), http.StatusBadRequest)
			return
		}

		// Extract userID and messageUUID from the ID field
		ids := strings.Split(emailData.ID, ":")
		if len(ids) != 2 {
			http.Error(w, "Invalid ID format, expected 'userID:messageUUID'", http.StatusBadRequest)
			return
		}

		// Check if userID and messageUUID are already set and match the current ones
		if userID != "" && userID != ids[0] {
			http.Error(w, "User ID mismatch", http.StatusBadRequest)
			return
		}
		if messageUUID != "" && messageUUID != ids[1] {
			http.Error(w, "Message UUID mismatch", http.StatusBadRequest)
			return
		}

		// Set the userID and messageUUID if they haven't been set yet
		if userID == "" {
			userID = ids[0]
		}
		if messageUUID == "" {
			messageUUID = ids[1]
		}

		// Create a new email using the constructor
		email := model.NewEmail(
			userID,
			messageUUID,
			emailData.Attributes.To,
			emailData.Attributes.Subject,
			emailData.Attributes.BodyHTML,
			emailData.Attributes.BodyText,
			emailData.Attributes.Attachments,
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

	// JSON:API-compliant response
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
			ID:   fmt.Sprintf("%s:%s", userID, messageUUID), // UserID:MessageUUID as ID
			Links: struct {
				Self string `json:"self"`
			}{
				Self: fmt.Sprintf("/email-queues/%s:%s", userID, messageUUID),
			},
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
