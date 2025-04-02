package eml

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	type caseStruct struct {
		name                string
		emlData             EML
		expectationFilePath string
	}

	testCases := []caseStruct{
		{
			name: "Valid",
			emlData: EML{
				MessageId:     "65ed6bfa-063c-5219-844d-e099c88a17f4",
				From:          "sender@test.multidialogo.it",
				ReplyTo:       "sender@test.multidialogo.it",
				To:            "recipient@test.multidialogo.it",
				Subject:       "Test Email",
				BodyHTML:      "<p>This is a test email in HTML format.</p>",
				BodyText:      "This is a test email in plain text format.",
				Date:          time.Unix(0, 0),
				Attachments:   []string{"testdata/resources/test_attachment.txt", "testdata/resources/doge.jpg"},
				CustomHeaders: map[string]string{"X-Custom-Header": "CustomHeaderValue"},
			},
			expectationFilePath: "testdata/expectations/65ed6bfa-063c-5219-844d-e099c88a17f4.EML",
		},
		{
			name: "Valid with Reply-To",
			emlData: EML{
				MessageId:     "ff0fb587-e29b-4278-bbab-a525196b8917",
				From:          "sender@test.multidialogo.it",
				ReplyTo:       "no-reply@test.multidialogo.it",
				To:            "recipient2@test.multidialogo.it",
				Subject:       "Test Email with Reply-To",
				BodyHTML:      "<p>This is another test email in HTML format.</p>",
				BodyText:      "This is another test email in plain text format.",
				Date:          time.Unix(0, 0),
				Attachments:   []string{},
				CustomHeaders: map[string]string{"X-Custom-Header": "AnotherHeaderValue"},
			},
			expectationFilePath: "testdata/expectations/ff0fb587-e29b-4278-bbab-a525196b8917.EML",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			sut := new(Writer)

			var actual bytes.Buffer
			err := sut.Write(&actual, tc.emlData)
			require.NoError(t, err)

			expected, _ := os.ReadFile(tc.expectationFilePath)
			assert.Equal(t, string(expected), actual.String())
		})
	}
}
