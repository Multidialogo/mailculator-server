package email

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func (m *databaseMock) Insert(_ context.Context, _ string, _ string) error {
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

func (m *databaseMock) GetStaleEmails(_ context.Context) ([]Email, error) {
	return nil, nil
}

func (m *databaseMock) GetInvalidEmails(_ context.Context) ([]Email, error) {
	return nil, nil
}

func (m *databaseMock) RequeueEmail(_ context.Context, _ string) error {
	return nil
}

func TestService_Save(t *testing.T) {
	t.Parallel()

	// Create EmailRequest batch with dummy payloads
	emailRequests := []EmailRequest{
		{
			MessageId:    "msg1",
			PayloadBytes: []byte("test payload 1"),
		},
		{
			MessageId:    "msg2",
			PayloadBytes: []byte("test payload 2"),
		},
	}

	type caseStruct struct {
		name                                   string
		payloadStorageErrorAfterCallCount      int
		databaseErrorAfterInsertCallCount      int
		expectPayloadStorageStoreCallCount     int
		expectedDatabaseInsertCallCount        int
		expectedDatabaseDeletePendingCallCount int
		expectError                            bool
	}

	testCases := []caseStruct{
		{
			name:                                   "success",
			payloadStorageErrorAfterCallCount:      2,
			databaseErrorAfterInsertCallCount:      2,
			expectPayloadStorageStoreCallCount:     2,
			expectedDatabaseInsertCallCount:        2,
			expectedDatabaseDeletePendingCallCount: 0,
			expectError:                            false,
		},
		{
			name:                                   "payload storage error",
			payloadStorageErrorAfterCallCount:      0,
			databaseErrorAfterInsertCallCount:      2,
			expectPayloadStorageStoreCallCount:     1,
			expectedDatabaseInsertCallCount:        0,
			expectedDatabaseDeletePendingCallCount: 0,
			expectError:                            true,
		},
		{
			name:                                   "database error",
			payloadStorageErrorAfterCallCount:      2,
			databaseErrorAfterInsertCallCount:      1,
			expectPayloadStorageStoreCallCount:     2,
			expectedDatabaseInsertCallCount:        2,
			expectedDatabaseDeletePendingCallCount: 1,
			expectError:                            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payloadStorage := &payloadStorageMock{errorAfterCallCount: tc.payloadStorageErrorAfterCallCount}
			database := &databaseMock{errorAfterInsertCallCount: tc.databaseErrorAfterInsertCallCount}

			sut := &Service{payloadStorage: payloadStorage, db: database}

			err := sut.Save(context.TODO(), emailRequests)
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
