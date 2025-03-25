package testutils

import (
	"fmt"
	"testing"
	"time"
)

// Named function to test GetCleanFunctionName
func DummyFunction() string {
	return GetCleanFunctionName()
}

// TestGetCleanFunctionName verifies that the clean function name is returned correctly
func TestGetCleanFunctionName(t *testing.T) {
	// Call the named test function and verify the result
	cleanName := DummyFunction()
	expectedName := "DummyFunction"

	if cleanName != expectedName {
		t.Errorf("expected %s, but got %s", expectedName, cleanName)
	}
}

// TestGetUnixEpoch verifies that GetUnixEpoch returns the correct Unix epoch time
func TestGetUnixEpoch(t *testing.T) {
	expectedTime := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

	// Call GetUnixEpoch and compare the result
	unixEpoch := GetUnixEpoch()
	if !unixEpoch.Equal(expectedTime) {
		t.Errorf("expected %v, but got %v", expectedTime, unixEpoch)
	}
}

func TestGenerateMessagePath(t *testing.T) {
	currentDate := time.Now()
	tests := []struct {
		name          string
		id            string
		expectedPath  string
		expectedError string
	}{
		{
			name:          "Valid ID format",
			id:            "65ed6bfa-063c-5219-844d-e099c88a17f4",
			expectedPath:  fmt.Sprintf("%d/%s/65ed6bfa-063c-5219-844d-e099c88a17f4", currentDate.Year(), currentDate.Month()),
			expectedError: "",
		},
		{
			name:          "Empty ID",
			id:            "",
			expectedPath:  "",
			expectedError: "invalid ID format: expected not empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GenerateMessagePath(tt.id)
			if err != nil && err.Error() != tt.expectedError {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if path != tt.expectedPath {
				t.Errorf("expected path: %v, got: %v", tt.expectedPath, path)
			}
		})
	}
}
