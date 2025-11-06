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

func (s *PayloadStorage) Store(messageId string, payload []byte) (string, error) {
	year, month, _ := time.Now().Date()
	dirPath := filepath.Join(s.basePath, fmt.Sprintf("%v/%v", year, month))
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dirPath, err)
	}

	filename := fmt.Sprintf("%s.json", messageId)
	path := filepath.Join(dirPath, filename)

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(payload); err != nil {
		return "", fmt.Errorf("failed to write payload: %w", err)
	}

	return path, nil
}
