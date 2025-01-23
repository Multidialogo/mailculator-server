package testutils

import (
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
