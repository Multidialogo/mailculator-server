package testutils

import (
	"runtime"
	"strings"
	"time"
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
