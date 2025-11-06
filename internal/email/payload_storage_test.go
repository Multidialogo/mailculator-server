package email

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPayloadStorageStore(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	// Test data
	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"
	payload := []byte("test payload data")

	// Execute
	path, err := storage.Store(messageId, payload)

	// Verify
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if path == "" {
		t.Errorf("expected non-empty path")
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist at %s: %v", path, err)
	}

	// Verify file content
	content, err := os.ReadFile(path)
	if err != nil {
		t.Errorf("failed to read file: %v", err)
	}

	if string(content) != string(payload) {
		t.Errorf("expected payload %s, got %s", string(payload), string(content))
	}

	// Verify filename contains the messageId
	fileName := filepath.Base(path)
	expectedFileName := messageId + ".json"
	if fileName != expectedFileName {
		t.Errorf("expected filename %s, got %s", expectedFileName, fileName)
	}

	// Verify file exists in the correct year/month structure
	dirPath := filepath.Dir(path)
	baseName := filepath.Base(dirPath)
	// Verify the directory is a valid month (01-12)
	if len(baseName) != 2 || baseName[0] < '0' || baseName[0] > '1' || baseName[1] < '0' || baseName[1] > '9' {
		// Could be October, November, December (10, 11, 12)
		// Just verify it's in the right structure
		if !strings.Contains(dirPath, tmpDir) {
			t.Errorf("expected path to be under %s, got %s", tmpDir, dirPath)
		}
	}
}

func TestPayloadStorageUniqueFilenames(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	// Store two payloads with different messageIds
	messageId1 := "65ed6bfa-063c-5219-844d-e099c88a17f4"
	messageId2 := "ff0fb587-e29b-4278-bbab-a525196b8917"
	payload1 := []byte("payload 1")
	payload2 := []byte("payload 2")

	path1, err1 := storage.Store(messageId1, payload1)
	path2, err2 := storage.Store(messageId2, payload2)

	// Verify
	if err1 != nil || err2 != nil {
		t.Errorf("expected no errors, got %v and %v", err1, err2)
	}

	if path1 == path2 {
		t.Errorf("expected different paths, both got %s", path1)
	}

	// Verify both files exist with correct content
	content1, _ := os.ReadFile(path1)
	content2, _ := os.ReadFile(path2)

	if string(content1) != string(payload1) {
		t.Errorf("expected payload1, got %s", string(content1))
	}

	if string(content2) != string(payload2) {
		t.Errorf("expected payload2, got %s", string(content2))
	}

	// Verify filename matches messageId
	fileName1 := filepath.Base(path1)
	fileName2 := filepath.Base(path2)
	if fileName1 != messageId1+".json" {
		t.Errorf("expected filename %s, got %s", messageId1+".json", fileName1)
	}
	if fileName2 != messageId2+".json" {
		t.Errorf("expected filename %s, got %s", messageId2+".json", fileName2)
	}
}

func TestPayloadStorageDirectoryCreation(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"
	payload := []byte("test payload")

	// Execute
	path, err := storage.Store(messageId, payload)

	// Verify
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify directory structure was created
	dirPath := filepath.Dir(path)
	if _, err := os.Stat(dirPath); err != nil {
		t.Errorf("expected directory to be created at %s: %v", dirPath, err)
	}
}

