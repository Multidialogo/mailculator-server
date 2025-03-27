package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator/v10"
	"log"
	"multicarrier-email-api/internal/API"
	"multicarrier-email-api/internal/config"
	"multicarrier-email-api/internal/model"
	"multicarrier-email-api/internal/outbox"
	"multicarrier-email-api/internal/service"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var emailQueueStorage *service.EmailQueueStorage

var inputPath string

var outboxService *outbox.Outbox

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

	// Initialize dynamo db outbox
	awsConfig := aws.Config{
		Region: registry.Get("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			registry.Get("AWS_ACCESS_KEY_ID"),
			registry.Get("AWS_SECRET_ACCESS_KEY"),
			"",
		),
		BaseEndpoint: aws.String(registry.Get("AWS_BASE_ENDPOINT")),
	}
	db := dynamodb.NewFromConfig(awsConfig)
	outboxService = outbox.NewOutbox(db)
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

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(APIRequest)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Process each email data in the request
	var emails []*model.Email
	var emailsOutbox []outbox.Email

	for _, emailData := range APIRequest.Data {

		attachmentPaths := make([]string, len(emailData.Attachments))
		for i, attachmentPath := range emailData.Attachments {
			attachmentPaths[i] = filepath.Join(inputPath, attachmentPath)
		}

		// Create a new email using the constructor
		email := model.NewEmail(
			emailData.ID,
			emailData.From,
			emailData.ReplyTo,
			emailData.To,
			emailData.Subject,
			emailData.BodyHTML,
			emailData.BodyText,
			attachmentPaths,
			emailData.CustomHeaders,
			time.Now(),
			emailData.CallbackCallOnSuccess,
			emailData.CallbackCallOnFailure,
		)

		// Append the created email to the slice
		emails = append(emails, email)

		emailsOutbox = append(
			emailsOutbox,
			outbox.Email{
				Id:              email.MessageUUID(),
				Status:          outbox.StatusProcessing,
				EmlFilePath:     email.Path(),
				SuccessCallback: email.CallbackCallOnSuccess(),
				FailureCallback: email.CallbackCallOnFailure(),
			},
		)
	}

	// Save the emails using the EmailQueueStorage service
	err = emailQueueStorage.SaveEmailsAsEML(emails)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving emails: %v", err), http.StatusInternalServerError)
		return
	}

	// storing data in the outbox table
	ctx := context.TODO()
	err = outboxService.BulkInsert(ctx, emailsOutbox)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error saving emails on db: %v", err), http.StatusInternalServerError)
		return
	}

	// Respond with the 201 Created status and the location of the saved emails
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/vnd.api+json") // Set the correct content type

	// JSON:APIRequest-compliant response
	response := struct {
		Data struct {
			Type string `json:"type"`
		} `json:"data"`
	}{
		Data: struct {
			Type string `json:"type"`
		}{
			Type: "mail-queue",
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding response: %v", err), http.StatusInternalServerError)
	}
}
