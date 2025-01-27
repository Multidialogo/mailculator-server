package API

import (
	"fmt"
	"regexp"
	"strings"
)

// Email validation regular expression
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// validateRequest validates the incoming request for required fields and checks for valid email addresses
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
		if datum.Attributes.From == "" {
			return fmt.Errorf("missing 'from' field")
		}
		if !isValidEmail(datum.Attributes.From) {
			return fmt.Errorf("invalid 'from' email address")
		}

		if datum.Attributes.To == "" {
			return fmt.Errorf("missing 'to' field")
		}
		if !isValidEmail(datum.Attributes.To) {
			return fmt.Errorf("invalid 'to' email address")
		}

		if datum.Attributes.ReplyTo == "" {
			return fmt.Errorf("missing 'replyTo' field")
		}
		if !isValidEmail(datum.Attributes.ReplyTo) {
			return fmt.Errorf("invalid 'replyTo' email address")
		}

		if datum.Attributes.Subject == "" {
			return fmt.Errorf("missing 'subject' field")
		}

		// If body HTML is empty, body text must also be empty
		if datum.Attributes.BodyHTML == "" && datum.Attributes.BodyText == "" {
			return fmt.Errorf("either 'bodyHTML' or 'bodyText' must be provided")
		}
	}

	return nil
}

// isValidEmail validates the email address using the regex
func isValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}
