package API

import (
	"fmt"
	"github.com/olesho/curl-parser"
	"regexp"
	"strings"
)

// Email validation regular expression
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateRequest validateRequest validates the incoming request for required fields and checks for valid email addresses
func ValidateRequest(APIRequest *QueueCreationAPI) error {
	if len(APIRequest.Data) == 0 {
		return fmt.Errorf("no email data provided")
	}

	for _, datum := range APIRequest.Data {
		// Check if the type is "email"
		if datum.Type != "email" {
			return fmt.Errorf("invalid type '%s', expected 'email'", datum.Type)
		}

		// Ensure ID is in the correct format
		ids := strings.Split(datum.ID, ":")
		if len(ids) != 3 {
			return fmt.Errorf("invalid ID format, expected 'userID:queueUUID:messageUUID'")
		}

		// Validate required fields
		if datum.From == "" {
			return fmt.Errorf("missing 'from' field")
		}
		if !isValidEmail(datum.From) {
			return fmt.Errorf("invalid 'from' email address")
		}

		if datum.To == "" {
			return fmt.Errorf("missing 'to' field")
		}
		if !isValidEmail(datum.To) {
			return fmt.Errorf("invalid 'to' email address")
		}

		if datum.ReplyTo == "" {
			return fmt.Errorf("missing 'reply_to' field")
		}
		if !isValidEmail(datum.ReplyTo) {
			return fmt.Errorf("invalid 'reply_to' email address")
		}

		if datum.Subject == "" {
			return fmt.Errorf("missing 'subject' field")
		}

		// If body HTML is empty, body text must also be empty
		if datum.BodyHTML == "" && datum.BodyText == "" {
			return fmt.Errorf("either 'body_html' or 'body_text' must be provided")
		}

		if datum.CallbackCallOnSuccess != "" {
			if !isValidCurlCommand(datum.CallbackCallOnSuccess) {
				return fmt.Errorf("invalid 'callback_on_success' curl command")
			}
		}

		if datum.CallbackCallOnFailure != "" {
			if !isValidCurlCommand(datum.CallbackCallOnFailure) {
				return fmt.Errorf("invalid 'callback_on_failure' curl command")
			}
		}
	}

	return nil
}

// isValidEmail validates the email address using the regex
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func isValidCurlCommand(curlCommand string) bool {
	if strings.Index(curlCommand, "curl ") != 0 {
		return false
	}
	_, err := parser.Parse(curlCommand)
	if err != nil {
		return false
	}
	return true
}
