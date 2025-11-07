package email

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"multicarrier-email-api/internal/email/testutils"
	"multicarrier-email-api/internal/eml"
)

type emlStorageMock struct {
	callCount           int
	errorAfterCallCount int
}

func (m *emlStorageMock) Store(_ eml.EML) (string, error) {
	m.callCount++

	if m.callCount > m.errorAfterCallCount {
		return "", errors.New("mock error")
	}

	return "file.EML", nil
}

type payloadStorageMock struct {
	callCount           int
	errorAfterCallCount int
}

func (m *payloadStorageMock) Store(_ string, _ []byte) (string, error) {
	m.callCount++

	if m.callCount > m.errorAfterCallCount {
		return "", errors.New("mock error")
	}

	return "payload_file", nil
}

type databaseMock struct {
	insertCallCount           int
	errorAfterInsertCallCount int
	deletedCallCount          int
}

func (m *databaseMock) Insert(_ context.Context, _ string, _ string, _ string) error {
	m.insertCallCount++

	if m.insertCallCount > m.errorAfterInsertCallCount {
		return errors.New("mock error")
	}

	return nil
}

func (m *databaseMock) DeletePending(_ context.Context, _ string) error {
	m.deletedCallCount++
	return nil
}

func (m *databaseMock) GetStaleEmails(_ context.Context) ([]StaleEmail, error) {
	return nil, nil
}

func (m *databaseMock) RequeueEmail(_ context.Context, _ string) error {
	return nil
}

func TestService_Save(t *testing.T) {
	t.Parallel()

	emlBatch := testutils.DummyEMLDataBatch(2)
	
	// Create EmailRequest batch with dummy payloads
	emailRequests := make([]EmailRequest, len(emlBatch))
	for i, eml := range emlBatch {
		emailRequests[i] = EmailRequest{
			EML:          eml,
			PayloadBytes: []byte("test payload " + string(rune(i))),
		}
	}

	type caseStruct struct {
		name                                   string
		emlStorageErrorAfterCallCount          int
		payloadStorageErrorAfterCallCount      int
		databaseErrorAfterInsertCallCount      int
		expectEMLStorageStoreCallCount         int
		expectPayloadStorageStoreCallCount     int
		expectedDatabaseInsertCallCount        int
		expectedDatabaseDeletePendingCallCount int
		expectError                            bool
	}

	testCases := []caseStruct{
		{
			name:                                   "success",
			emlStorageErrorAfterCallCount:          2,
			payloadStorageErrorAfterCallCount:      2,
			databaseErrorAfterInsertCallCount:      2,
			expectEMLStorageStoreCallCount:         2,
			expectPayloadStorageStoreCallCount:     2,
			expectedDatabaseInsertCallCount:        2,
			expectedDatabaseDeletePendingCallCount: 0,
			expectError:                            false,
		},
		{
			name:                                   "eml storage error",
			emlStorageErrorAfterCallCount:          0,
			payloadStorageErrorAfterCallCount:      2,
			databaseErrorAfterInsertCallCount:      2,
			expectEMLStorageStoreCallCount:         1,
			expectPayloadStorageStoreCallCount:     0,
			expectedDatabaseInsertCallCount:        0,
			expectedDatabaseDeletePendingCallCount: 0,
			expectError:                            true,
		},
		{
			name:                                   "payload storage error",
			emlStorageErrorAfterCallCount:          2,
			payloadStorageErrorAfterCallCount:      1,
			databaseErrorAfterInsertCallCount:      2,
			expectEMLStorageStoreCallCount:         2,
			expectPayloadStorageStoreCallCount:     2,
			expectedDatabaseInsertCallCount:        1,
			expectedDatabaseDeletePendingCallCount: 1,
			expectError:                            true,
		},
		{
			name:                                   "database error",
			emlStorageErrorAfterCallCount:          2,
			payloadStorageErrorAfterCallCount:      2,
			databaseErrorAfterInsertCallCount:      1,
			expectEMLStorageStoreCallCount:         2,
			expectPayloadStorageStoreCallCount:     2,
			expectedDatabaseInsertCallCount:        2,
			expectedDatabaseDeletePendingCallCount: 1,
			expectError:                            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			emlStorage := &emlStorageMock{errorAfterCallCount: tc.emlStorageErrorAfterCallCount}
			payloadStorage := &payloadStorageMock{errorAfterCallCount: tc.payloadStorageErrorAfterCallCount}
			database := &databaseMock{errorAfterInsertCallCount: tc.databaseErrorAfterInsertCallCount}

			sut := &Service{emlStorage: emlStorage, payloadStorage: payloadStorage, db: database}

			err := sut.Save(context.TODO(), emailRequests)
			assert.Equal(t, tc.expectEMLStorageStoreCallCount, emlStorage.callCount)
			assert.Equal(t, tc.expectPayloadStorageStoreCallCount, payloadStorage.callCount)
			assert.Equal(t, tc.expectedDatabaseInsertCallCount, database.insertCallCount)
			assert.Equal(t, tc.expectedDatabaseDeletePendingCallCount, database.deletedCallCount)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
