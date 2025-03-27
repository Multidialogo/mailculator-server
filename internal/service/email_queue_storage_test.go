package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"multicarrier-email-api/internal/model"
	"multicarrier-email-api/internal/testutils"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

var fixturesDir string
var expectationsDir string

func init() {
	// Get the directory where the test source is located (i.e., the directory of this test file)
	_, currentFilePath, _, _ := runtime.Caller(0)
	testDir := filepath.Join(filepath.Dir(currentFilePath), "testData")
	fixturesDir = filepath.Join(testDir, "fixtures")
	expectationsDir = filepath.Join(testDir, "expectations")
}

func TestEmailQueueStorage_SaveEmailsAsEML(t *testing.T) {
	// Setup: Create a temporary test directory for the test
	basePath := t.TempDir() // t.TempDir() automatically creates a temp directory
	require.NotEmpty(t, basePath, "Temp dir should not be empty")

	draftPath := filepath.Join(basePath, "outbox")

	// Initialize EmailQueueStorage with the base path for storing EML files
	emailQueueStorage := NewEmailQueueStorage(
		filepath.Join(basePath, "outbox"),
	)

	// Define the test cases (data provider)
	tests := []struct {
		name                string
		email               *model.Email
		expectedEMLPath     string
		expectedEMLFileName string
	}{
		{
			name: "Valid email",
			email: model.NewEmail(
				"65ed6bfa-063c-5219-844d-e099c88a17f4",
				"sender@test.multidialogo.it",
				"sender@test.multidialogo.it",
				"recipient@test.multidialogo.it",
				"Test Email",
				"<p>This is a test email in HTML format.</p>",
				"This is a test email in plain text format.",
				[]string{
					filepath.Join(fixturesDir, testutils.GetCleanFunctionName(), "test_attachment.txt"),
					filepath.Join(fixturesDir, testutils.GetCleanFunctionName(), "doge.jpg"),
				},
				map[string]string{
					"X-Custom-Header": "CustomHeaderValue",
				},
				testutils.GetUnixEpoch(),
				"",
				"",
			),
			expectedEMLPath:     "1970/January/65ed6bfa-063c-5219-844d-e099c88a17f4.EML",
			expectedEMLFileName: "65ed6bfa-063c-5219-844d-e099c88a17f4.EML",
		},
		{
			name: "Valid email with Reply-To",
			email: model.NewEmail(
				"ff0fb587-e29b-4278-bbab-a525196b8917",
				"sender@test.multidialogo.it",
				"no-reply@test.multidialogo.it",
				"recipient2@test.multidialogo.it",
				"Test Email with Reply-To",
				"<p>This is another test email in HTML format.</p>",
				"This is another test email in plain text format.",
				[]string{},
				map[string]string{
					"X-Custom-Header": "AnotherHeaderValue",
				},
				testutils.GetUnixEpoch(),
				"",
				"",
			),
			expectedEMLPath:     "1970/January/ff0fb587-e29b-4278-bbab-a525196b8917.EML",
			expectedEMLFileName: "ff0fb587-e29b-4278-bbab-a525196b8917.EML",
		},
	}

	currentTestExpectationsDir := filepath.Join(expectationsDir, testutils.GetCleanFunctionName())

	// Execute the test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save the email as an EML file
			err := emailQueueStorage.SaveEmailsAsEML([]*model.Email{tt.email})
			require.NoError(t, err, "Failed to save email as EML")

			// Verify that the EML file was created
			actualEmlFilePath := filepath.Join(draftPath, tt.expectedEMLPath)
			_, err = os.Stat(actualEmlFilePath)
			require.NoError(t, err, "EML file was not created")

			// Read the contents of the EML file
			actualEmlFileContent, err := os.ReadFile(actualEmlFilePath)
			require.NoError(t, err, "Failed to read EML file")

			// Read expected EML content
			expectationEmlFileContent, err := os.ReadFile(filepath.Join(currentTestExpectationsDir, tt.expectedEMLFileName))
			require.NoError(t, err, "Failed to read expectation EML file")

			// Compare the contents
			assert.Equal(t, string(expectationEmlFileContent), string(actualEmlFileContent))
		})
	}
}
