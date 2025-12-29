package email

import (
	"context"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
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

func (m *payloadStorageMock) Delete(_ string) error {
	return nil
}

type databaseMock struct {
	insertCallCount           int
	errorAfterInsertCallCount int
	insertError               error
}

func (m *databaseMock) Insert(_ context.Context, _ string, _ string) error {
	m.insertCallCount++

	if m.insertCallCount > m.errorAfterInsertCallCount {
		if m.insertError != nil {
			return m.insertError
		}
		return errors.New("mock error")
	}

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
		name                               string
		payloadStorageErrorAfterCallCount  int
		databaseErrorAfterInsertCallCount  int
		expectPayloadStorageStoreCallCount int
		expectedDatabaseInsertCallCount    int
		expectedSuccessCount               int
		expectedFailCount                  int
	}

	testCases := []caseStruct{
		{
			name:                               "all succeed",
			payloadStorageErrorAfterCallCount:  2,
			databaseErrorAfterInsertCallCount:  2,
			expectPayloadStorageStoreCallCount: 2,
			expectedDatabaseInsertCallCount:    2,
			expectedSuccessCount:               2,
			expectedFailCount:                  0,
		},
		{
			name:                               "payload storage error on first",
			payloadStorageErrorAfterCallCount:  0,
			databaseErrorAfterInsertCallCount:  2,
			expectPayloadStorageStoreCallCount: 2,
			expectedDatabaseInsertCallCount:    0,
			expectedSuccessCount:               0,
			expectedFailCount:                  2,
		},
		{
			name:                               "payload storage error on second",
			payloadStorageErrorAfterCallCount:  1,
			databaseErrorAfterInsertCallCount:  2,
			expectPayloadStorageStoreCallCount: 2,
			expectedDatabaseInsertCallCount:    1,
			expectedSuccessCount:               1,
			expectedFailCount:                  1,
		},
		{
			name:                               "database error on first",
			payloadStorageErrorAfterCallCount:  2,
			databaseErrorAfterInsertCallCount:  0,
			expectPayloadStorageStoreCallCount: 2,
			expectedDatabaseInsertCallCount:    2,
			expectedSuccessCount:               0,
			expectedFailCount:                  2,
		},
		{
			name:                               "database error on second",
			payloadStorageErrorAfterCallCount:  2,
			databaseErrorAfterInsertCallCount:  1,
			expectPayloadStorageStoreCallCount: 2,
			expectedDatabaseInsertCallCount:    2,
			expectedSuccessCount:               1,
			expectedFailCount:                  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payloadStorage := &payloadStorageMock{errorAfterCallCount: tc.payloadStorageErrorAfterCallCount}
			database := &databaseMock{errorAfterInsertCallCount: tc.databaseErrorAfterInsertCallCount}

			sut := &Service{payloadStorage: payloadStorage, db: database}

			results := sut.Save(context.TODO(), emailRequests)

			assert.Equal(t, len(emailRequests), len(results))
			assert.Equal(t, tc.expectPayloadStorageStoreCallCount, payloadStorage.callCount)
			assert.Equal(t, tc.expectedDatabaseInsertCallCount, database.insertCallCount)

			// Count successes and failures
			successCount := 0
			failCount := 0
			for _, result := range results {
				if result.Success {
					successCount++
				} else {
					failCount++
					assert.NotEmpty(t, result.ErrorCode)
				}
			}

			assert.Equal(t, tc.expectedSuccessCount, successCount)
			assert.Equal(t, tc.expectedFailCount, failCount)
		})
	}
}

func TestService_Save_ErrorTypes(t *testing.T) {
	t.Parallel()

	emailRequests := []EmailRequest{
		{
			MessageId:    "msg1",
			PayloadBytes: []byte("test payload 1"),
		},
	}

	testCases := []struct {
		name              string
		dbError           error
		expectedErrorCode string
	}{
		{
			name:              "duplicate entry error",
			dbError:           &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			expectedErrorCode: ErrorCodeDuplicatedID,
		},
		{
			name:              "generic database error",
			dbError:           errors.New("generic error"),
			expectedErrorCode: ErrorCodeDatabaseError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payloadStorage := &payloadStorageMock{errorAfterCallCount: 1}
			database := &databaseMock{
				errorAfterInsertCallCount: 0,
				insertError:               tc.dbError,
			}

			sut := &Service{payloadStorage: payloadStorage, db: database}

			results := sut.Save(context.TODO(), emailRequests)

			assert.Len(t, results, 1)
			assert.False(t, results[0].Success)
			assert.Equal(t, tc.expectedErrorCode, results[0].ErrorCode)
			assert.NotEmpty(t, results[0].ErrorMessage)
		})
	}
}
