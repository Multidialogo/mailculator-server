package service

import (
	"os"
	"testing"
	"mailculator/internal/model"
	"mailculator/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"runtime"
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

	// Initialize EmailQueueStorage with the base path for storing EML files
	emailQueueStorage := NewEmailQueueStorage(basePath)

	// Create a sample email model
	email := model.NewEmail(
		"user1",
		"queue1",
		"message1",
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
	)

	// Save the email as an EML file
	err := emailQueueStorage.SaveEmailsAsEML([]*model.Email{email})
	require.NoError(t, err, "Failed to save email as EML")

	// Verify that the EML file was created
	emlFilePath := filepath.Join(basePath, "users/user1/queues/queue1/messages/message1.EML")
	_, err = os.Stat(emlFilePath)
	require.NoError(t, err, "EML file was not created")

	// Read the contents of the EML file
	emlFileContent, err := os.ReadFile(emlFilePath)
	require.NoError(t, err, "Failed to read EML file")

	//err = os.WriteFile(filepath.Join(expectationsDir, testutils.GetCleanFunctionName(), "user1:message1.EML"), []byte(emlFileContent), 0644)

	expectationEmlFileContent, err := os.ReadFile(filepath.Join(expectationsDir, testutils.GetCleanFunctionName(), "user1:queue1:message1.EML"))
	require.NoError(t, err, "Failed to read expectation EML file")

	assert.Equal(t, string(expectationEmlFileContent), string(emlFileContent))
}
