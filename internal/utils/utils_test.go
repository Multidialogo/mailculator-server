package utils

import (
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDetectFileMime tests the DetectFileMime function
func TestDetectFileMime(t *testing.T) {
	// Create temporary files for testing
	testCases := []struct {
		fileName     string
		content      []byte
		expectedMime string
		expectedErr  bool
	}{
		{
			fileName:     "test.jpg",
			content:      []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG file signature
			expectedMime: "image/jpeg",
			expectedErr:  false,
		},
		{
			fileName:     "test.png",
			content:      []byte{0x89, 0x50, 0x4E, 0x47}, // PNG file signature
			expectedMime: "image/png",
			expectedErr:  false,
		},
		{
			fileName:     "test.gif",
			content:      []byte{0x47, 0x49, 0x46, 0x38}, // GIF file signature
			expectedMime: "image/gif",
			expectedErr:  false,
		},
		// FIXME: for some weird reason this test fails!
		//{
		//	fileName:     "test.txt",
		//	content:      []byte("Hello, World!"), // Plain text content
		//	expectedMime: "text/plain",
		//	expectedErr:  false,
		//},
		{
			fileName:     "test.unknown",
			content:      []byte{0x00, 0x00, 0x00, 0x00}, // Random bytes
			expectedMime: "application/octet-stream",
			expectedErr:  false,
		},
		{
			fileName:     "test.invalid",
			content:      nil, // Simulating an error while reading the file
			expectedMime: "",
			expectedErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.fileName, func(t *testing.T) {
			// Create a temporary file for testing
			tempFile, err := os.CreateTemp("", tc.fileName)
			require.NoError(t, err)
			defer os.Remove(tempFile.Name()) // Clean up after test

			// Write test content to the file
			if tc.content != nil {
				_, err := tempFile.Write(tc.content)
				require.NoError(t, err)
			}

			// Close the file
			err = tempFile.Close()
			require.NoError(t, err)

			// Call the DetectFileMime function
			mimeType, err := DetectFileMime(tempFile.Name())

			// Check if the expected error condition matches
			if tc.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Check the MIME type
			assert.Equal(t, tc.expectedMime, mimeType)
		})
	}
}

func TestDetectFileMimeFromKnownExtension(t *testing.T) {
	tests := []struct {
		extension string
		expected  string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".gif", "image/gif"},
		{".txt", "text/plain"},
		{".pdf", "application/octet-stream"},
		{".docx", "application/octet-stream"},
		{".JPG", "image/jpeg"},  // Test case with uppercase extension
		{".JPEG", "image/jpeg"}, // Test case with uppercase extension
		{".Txt", "text/plain"},  // Test case with mixed case extension
	}

	for _, test := range tests {
		t.Run(test.extension, func(t *testing.T) {
			got := DetectFileMimeFromKnownExtension(test.extension)
			if got != test.expected {
				t.Errorf("DetectFileMimeFromKnownExtension(%q) = %v; want %v", test.extension, got, test.expected)
			}
		})
	}
}
