package testutils

import (
	"runtime"
	"strings"
	"time"
	"os"
	"path/filepath"
	"testing"
	"io"
	"fmt"
)

// GetCleanFunctionName returns the clean function name without path qualification
func GetCleanFunctionName() string {
	// Get the program counter and then the function name
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()

	// Extract the clean function name without path qualification
	// Split by the last '.' to remove package/module part
	parts := strings.Split(funcName, ".")
	return parts[len(parts)-1]
}

func GetUnixEpoch() time.Time {
	// Parse the date string into a time.Time object
	dateString := "Thu, 01 Jan 1970 00:00:00 +0000"
	layout := time.RFC1123Z
	parsedTime, err := time.Parse(layout, dateString)
	if err != nil {
		panic(err) // Handle the error properly in production code
	}

	return parsedTime
}

func LoadFixturesFilesInInputDirectory(testPayloadDir string, testInputDir string, t *testing.T) {
	err := os.MkdirAll(testInputDir, os.ModePerm)
	if err != nil {
		t.Fatalf("Failed to create test files directory \"%s\": %v", testInputDir, err)
	}

	err = filepath.Walk(testPayloadDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a regular file (not a directory)
		if info.IsDir() {
			return nil
		}

		// Open the source file
		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Define the destination path in /tmp
		dstPath := filepath.Join(testInputDir, info.Name())

		// Create the destination file
		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		// Copy the content of the source file to the destination file
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Error walking through directory: %v", err)
	}
}

// GenerateMessagePath generates the path based on the ID format 'userID:queueUUID:messageUUID'
func GenerateMessagePath(id string) (string, error) {
	// Split the ID into parts
	parts := strings.Split(id, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid ID format: expected 'userID:queueUUID:messageUUID'")
	}

	// Extract userID, queueUUID, and messageUUID
	userID := parts[0]
	queueUUID := parts[1]
	messageUUID := parts[2]

	// Check if any part is empty
	if userID == "" || queueUUID == "" || messageUUID == "" {
		return "", fmt.Errorf("invalid ID format: expected 'userID:queueUUID:messageUUID'")
	}

	// Generate the new path
	path := fmt.Sprintf("users/%s/queues/%s/messages/%s", userID, queueUUID, messageUUID)
	return path, nil
}
