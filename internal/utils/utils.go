package utils

import (
	"fmt"
	"github.com/h2non/filetype"
	"os"
	"path/filepath"
	"strings"
	"io"
)

// DetectFileMime detects the MIME type of a file using file signature and extension
func DetectFileMime(path string) (string, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("Error opening file: %w", err)
	}
	defer file.Close()

	// Read the first few bytes to detect content type using filetype
	buffer := make([]byte, 261) // Read first 261 bytes, larger buffer for better detection
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("Error reading file: %w", err)
	}

	// Use the filetype package to detect the MIME type based on file signature
	kind, _ := filetype.Match(buffer)
	if kind == filetype.Unknown || kind.MIME.Value == "application/octet-stream" {
		// Fallback to extension-based detection for known types
		return DetectFileMimeFromKnownExtension(filepath.Ext(path)), nil
	}

	// If MIME type is found via file signature, return it
	return kind.MIME.Value, nil
}

func DetectFileMimeFromKnownExtension(extension string) string {
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// CopyFile copies a file from src to dest, ensuring the destination path exists
func CopyFile(src, dest string) error {
	// Ensure the destination directory exists
	destDir := filepath.Dir(dest)
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	destFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy the contents of the source file to the destination file
	if _, err := io.Copy(destFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Optionally, set the destination file permissions to match the source file
	srcInfo, err := os.Stat(src)
	if err == nil {
		if err := os.Chmod(dest, srcInfo.Mode()); err != nil {
			return fmt.Errorf("failed to set file permissions: %w", err)
		}
	}

	return nil
}
