package service

import (
	"fmt"
	"mailculator/internal/model"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"os"
	"path/filepath"
	"time"
	"io/ioutil"
	"encoding/base64"
	"net/textproto"
	"github.com/h2non/filetype"
)

type EmailQueueStorage struct {
	DraftOutputPath string
}

func NewEmailQueueStorage(draftOutputPath string) *EmailQueueStorage {
	return &EmailQueueStorage{DraftOutputPath: draftOutputPath}
}

func (s *EmailQueueStorage) SaveEmailsAsEML(emails []*model.Email) error {
	for _, email := range emails {
		// Generate file path for the .EML file
		filePath := filepath.Join(s.DraftOutputPath, fmt.Sprintf("%s.EML", email.Path()))

		// Ensure the directory structure exists
		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directories for EML file: %w", err)
		}

		// Open the file for writing
		file, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create EML file: %w", err)
		}
		defer file.Close()

		// Create the MIME header and body
		msg := &mail.Message{}
		addHeadersToMessage(msg, email)

		// Write the standard headers
		orderedStandardHeaders := []string{"From", "Reply-To", "To", "Date", "Subject", "Content-Type"}

		// Write standard headers
		for _, key := range orderedStandardHeaders {
			if values, exists := msg.Header[key]; exists {
				for _, value := range values {
					_, err := file.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
					if err != nil {
						return fmt.Errorf("failed to write header %s: %w", key, err)
					}
				}
			}
		}

		// Write custom headers (those not in the standard order)
		for key, values := range msg.Header {
			// Skip headers that were already written in the standard order
			if isHeaderInList(orderedStandardHeaders, key) {
				continue
			}
			for _, value := range values {
				_, err := file.Write([]byte(fmt.Sprintf("%s: %s\r\n", key, value)))
				if err != nil {
					return fmt.Errorf("failed to write custom header %s: %w", key, err)
				}
			}
		}

		// Write a newline after the custom headers to separate them from the body
		_, err = file.Write([]byte("\r\n"))
		if err != nil {
			return fmt.Errorf("failed to write newline after custom headers: %w", err)
		}

		// Write body and any attachments to the EML file
		multipartWriter := multipart.NewWriter(file)
		multipartWriter.SetBoundary(email.MessageUUID())

		// Write the plain-text body
		if email.BodyText() != "" {
			err = writePart(multipartWriter, "text/plain", "charset=utf-8", email.BodyText())
			if err != nil {
				return err
			}
		}

		// Write the HTML body
		if email.BodyHTML() != "" {
			err = writePart(multipartWriter, "text/html", "charset=utf-8", email.BodyHTML())
			if err != nil {
				return err
			}
		}

		// Write attachments if any
		for _, attachment := range email.Attachments() { // Call Attachments if it's a function
			// Read the attachment file
			attachmentData, err := ioutil.ReadFile(attachment)
			if err != nil {
				return fmt.Errorf("failed to read attachment: %w", err)
			}
			err = writeAttachment(multipartWriter, attachment, attachmentData)
			if err != nil {
				return err
			}
		}

		// Close the multipart writer to finish writing the body and attachments
		err = multipartWriter.Close()
		if err != nil {
			return fmt.Errorf("failed to close multipart writer: %w", err)
		}
	}

	return nil
}

// Helper function to write email part (text or HTML)
func writePart(multipartWriter *multipart.Writer, contentType, charset, body string) error {
	// Convert the header to MIMEHeader
	headers := textproto.MIMEHeader{
		"Content-Type":              []string{fmt.Sprintf("%s; %s", contentType, charset)},
		"Content-Transfer-Encoding": []string{"quoted-printable"},
	}

	part, err := multipartWriter.CreatePart(headers)
	if err != nil {
		return fmt.Errorf("failed to create part: %w", err)
	}

	writer := quotedprintable.NewWriter(part)
	_, err = writer.Write([]byte(body))
	if err != nil {
		return fmt.Errorf("failed to write part body: %w", err)
	}
	writer.Close()
	return nil
}

// Helper function to write an attachment part
func writeAttachment(multipartWriter *multipart.Writer, path string, data []byte) error {
	mimeType, err := detectFileMime(path)
	if err != nil {
		return fmt.Errorf("failed to detect file mime type: %w", err)
	}

	// Create the part for the attachment
	headers := textproto.MIMEHeader{
		"Content-Type":              []string{mimeType},
		"Content-Disposition":       []string{fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(path))},
		"Content-Transfer-Encoding": []string{"base64"},
	}

	part, err := multipartWriter.CreatePart(headers)
	if err != nil {
		return fmt.Errorf("failed to create attachment part: %w", err)
	}

	// Encode the attachment in base64 and write to the part
	base64Encoder := base64.NewEncoder(base64.StdEncoding, part)
	_, err = base64Encoder.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write attachment data: %w", err)
	}
	base64Encoder.Close()
	return nil
}

// detectFileMime detects the MIME type of a file using file signature and extension
func detectFileMime(path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Error opening file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes to detect content type using filetype
	buffer := make([]byte, 261) // Read first 261 bytes, larger buffer for better detection
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("Error reading file: %w", err)
	}

	// Use the filetype package to detect the MIME type based on file signature
	if kind, _ := filetype.Match(buffer); kind != filetype.Unknown {
		return kind.MIME.Value, nil
	}

	// Fallback to extension-based detection for known types
	ext := filepath.Ext(path)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg", nil
	case ".png":
		return "image/png", nil
	case ".gif":
		return "image/gif", nil
	case ".txt":
		return "text/plain", nil
	}

	// Return application/octet-stream if no MIME type was found
	return "application/octet-stream", nil
}

func isHeaderInList(slice []string, item string) bool {
	for _, element := range slice {
		if element == item {
			return true
		}
	}
	return false
}

// AddHeadersToMessage adds standard and custom headers to the given message
func addHeadersToMessage(msg *mail.Message, email *model.Email) {
	// Set standard email headers
	msg.Header = make(mail.Header)
	msg.Header["From"] = []string{email.From()}
	// Set Reply-To header only if it differs from From
	if email.ReplyTo() != email.From() {
		msg.Header["Reply-To"] = []string{email.ReplyTo()}
	}
	msg.Header["To"] = []string{email.To()}
	msg.Header["Date"] = []string{email.Date().Format(time.RFC1123Z)}
	msg.Header["Subject"] = []string{email.Subject()}
	msg.Header["Content-Type"] = []string{fmt.Sprintf("multipart/mixed; boundary=\"%s\"", email.MessageUUID())}

	for key, value := range email.CustomHeaders() {
		msg.Header[key] = []string{value}
	}
}
