package email

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type PayloadStorage struct {
	basePath string
}

func NewPayloadStorage(basePath string) *PayloadStorage {
	return &PayloadStorage{basePath}
}

func (s *PayloadStorage) ResolvePath(messageId string) (string, error) {
	year, month, _ := time.Now().Date()
	dirPath := filepath.Join(s.basePath, fmt.Sprintf("%v/%v", year, month))
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	return filepath.Join(dirPath, fmt.Sprintf("%s.json", messageId)), nil
}

func (s *PayloadStorage) Write(path string, payload []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(payload); err != nil {
		return fmt.Errorf("failed to write payload: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync payload to disk: %w", err)
	}

	return nil
}

func (s *PayloadStorage) Delete(payloadPath string) error {
	if err := os.Remove(payloadPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete payload file %s: %w", payloadPath, err)
	}
	return nil
}
