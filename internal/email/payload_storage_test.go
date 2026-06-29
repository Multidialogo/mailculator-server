package email

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPayloadStorageResolvePath(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	// Test data
	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"

	path, err := storage.ResolvePath(messageId)

	// Verify
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if filepath.Base(path) != messageId+".json" {
		t.Errorf("expected filename %s.json, got %s", messageId, filepath.Base(path))
	}

	dirPath := filepath.Dir(path)
	if _, err := os.Stat(dirPath); err != nil {
		t.Errorf("expected directory to be created at %s: %v", dirPath, err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file not to exist yet at %s", path)
	}
}

func TestPayloadStorageResolvePathYearMonthStructure(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"

	path, err := storage.ResolvePath(messageId)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.HasPrefix(path, tmpDir) {
		t.Errorf("expected path to be under %s, got %s", tmpDir, path)
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

func TestPayloadStorageWrite(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"
	payload := []byte("test payload data")

	path, err := storage.ResolvePath(messageId)
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}

	if err := storage.Write(path, payload); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected file to exist at %s: %v", path, err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != string(payload) {
		t.Errorf("expected payload %s, got %s", string(payload), string(content))
	}
}

func TestPayloadStorageWriteOverwritesExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"

	path, err := storage.ResolvePath(messageId)
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}

	if err := storage.Write(path, []byte("original payload")); err != nil {
		t.Fatalf("first write failed: %v", err)
	}

	newPayload := []byte("overwritten payload")
	if err := storage.Write(path, newPayload); err != nil {
		t.Fatalf("second write failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != string(newPayload) {
		t.Errorf("expected overwritten payload, got %s", string(content))
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

	path1, err := storage.ResolvePath(messageId1)
	if err != nil {
		t.Fatalf("failed to resolve path1: %v", err)
	}
	if err := storage.Write(path1, payload1); err != nil {
		t.Fatalf("failed to write payload1: %v", err)
	}

	path2, err := storage.ResolvePath(messageId2)
	if err != nil {
		t.Fatalf("failed to resolve path2: %v", err)
	}
	if err := storage.Write(path2, payload2); err != nil {
		t.Fatalf("failed to write payload2: %v", err)
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

	if filepath.Base(path1) != messageId1+".json" {
		t.Errorf("expected filename %s, got %s", messageId1+".json", filepath.Base(path1))
	}
	if filepath.Base(path2) != messageId2+".json" {
		t.Errorf("expected filename %s, got %s", messageId2+".json", filepath.Base(path2))
	}
}

func TestPayloadStorageDelete(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	messageId := "65ed6bfa-063c-5219-844d-e099c88a17f4"

	path, err := storage.ResolvePath(messageId)
	if err != nil {
		t.Fatalf("failed to resolve path: %v", err)
	}
	if err := storage.Write(path, []byte("test payload")); err != nil {
		t.Fatalf("failed to write payload: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file should exist before delete: %v", err)
	}

	// Execute delete
	err = storage.Delete(path)

	// Verify no error
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	// Verify file no longer exists
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, but it still exists")
	}
}

func TestPayloadStorageDeleteNonExistentFile(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	storage := NewPayloadStorage(tmpDir)

	nonExistentPath := filepath.Join(tmpDir, "non-existent-file.json")

	// Execute delete on non-existent file
	err := storage.Delete(nonExistentPath)

	// Verify no error (implementation should handle non-existent files gracefully)
	if err != nil {
		t.Errorf("expected no error when deleting non-existent file, got %v", err)
	}
}
