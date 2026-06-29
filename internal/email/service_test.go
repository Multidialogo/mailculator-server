package email

import (
	"context"
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

type payloadStorageMock struct {
	resolvePathCallCount           int
	resolvePathErrorAfterCallCount int
	writeCallCount                 int
	writeErrorAfterCallCount       int
	deleteCallCount                int
}

func (m *payloadStorageMock) ResolvePath(_ string) (string, error) {
	m.resolvePathCallCount++
	if m.resolvePathCallCount > m.resolvePathErrorAfterCallCount {
		return "", errors.New("mock resolve path error")
	}
	return "payload_file", nil
}

func (m *payloadStorageMock) Write(_ string, _ []byte) error {
	m.writeCallCount++
	if m.writeCallCount > m.writeErrorAfterCallCount {
		return errors.New("mock write error")
	}
	return nil
}

func (m *payloadStorageMock) Delete(_ string) error {
	m.deleteCallCount++
	return nil
}

type databaseMock struct {
	insertCallCount           int
	errorAfterInsertCallCount int
	insertError               error
	postCallbackError         error
}

func (m *databaseMock) Insert(_ context.Context, _ string, _ string, onEmailInserted func() error) error {
	m.insertCallCount++

	if m.insertCallCount > m.errorAfterInsertCallCount {
		if m.insertError != nil {
			return m.insertError
		}
		return errors.New("mock error")
	}

	if onEmailInserted != nil {
		if err := onEmailInserted(); err != nil {
			return err
		}
	}

	if m.postCallbackError != nil {
		return &commitError{err: m.postCallbackError}
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
		name                              string
		resolvePathErrorAfterCallCount    int
		writeErrorAfterCallCount          int
		databaseErrorAfterInsertCallCount int
		expectedResolvePathCallCount      int
		expectedWriteCallCount            int
		expectedDatabaseInsertCallCount   int
		expectedSuccessCount              int
		expectedFailCount                 int
	}

	testCases := []caseStruct{
		{
			name:                              "all succeed",
			resolvePathErrorAfterCallCount:    2,
			writeErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount: 2,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            2,
			expectedDatabaseInsertCallCount:   2,
			expectedSuccessCount:              2,
			expectedFailCount:                 0,
		},
		{
			name:                              "resolve path error on all",
			resolvePathErrorAfterCallCount:    0,
			writeErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount: 2,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            0,
			expectedDatabaseInsertCallCount:   0,
			expectedSuccessCount:              0,
			expectedFailCount:                 2,
		},
		{
			name:                              "resolve path error on second",
			resolvePathErrorAfterCallCount:    1,
			writeErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount: 2,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            1,
			expectedDatabaseInsertCallCount:   1,
			expectedSuccessCount:              1,
			expectedFailCount:                 1,
		},
		{
			name:                              "write error on all",
			resolvePathErrorAfterCallCount:    2,
			writeErrorAfterCallCount:          0,
			databaseErrorAfterInsertCallCount: 2,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            2,
			expectedDatabaseInsertCallCount:   2,
			expectedSuccessCount:              0,
			expectedFailCount:                 2,
		},
		{
			name:                              "write error on second",
			resolvePathErrorAfterCallCount:    2,
			writeErrorAfterCallCount:          1,
			databaseErrorAfterInsertCallCount: 2,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            2,
			expectedDatabaseInsertCallCount:   2,
			expectedSuccessCount:              1,
			expectedFailCount:                 1,
		},
		{
			name:                              "database error on all",
			resolvePathErrorAfterCallCount:    2,
			writeErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount: 0,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            0,
			expectedDatabaseInsertCallCount:   2,
			expectedSuccessCount:              0,
			expectedFailCount:                 2,
		},
		{
			name:                              "database error on second",
			resolvePathErrorAfterCallCount:    2,
			writeErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount: 1,
			expectedResolvePathCallCount:      2,
			expectedWriteCallCount:            1,
			expectedDatabaseInsertCallCount:   2,
			expectedSuccessCount:              1,
			expectedFailCount:                 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payloadStorage := &payloadStorageMock{
				resolvePathErrorAfterCallCount: tc.resolvePathErrorAfterCallCount,
				writeErrorAfterCallCount:       tc.writeErrorAfterCallCount,
			}
			database := &databaseMock{errorAfterInsertCallCount: tc.databaseErrorAfterInsertCallCount}

			sut := &Service{payloadStorage: payloadStorage, db: database}

			results := sut.Save(context.TODO(), emailRequests)

			assert.Equal(t, len(emailRequests), len(results))
			assert.Equal(t, tc.expectedResolvePathCallCount, payloadStorage.resolvePathCallCount)
			assert.Equal(t, tc.expectedWriteCallCount, payloadStorage.writeCallCount)
			assert.Equal(t, tc.expectedDatabaseInsertCallCount, database.insertCallCount)

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
		name                           string
		dbErrorAfterInsertCallCount    int
		dbError                        error
		writeErrorAfterCallCount       int
		resolvePathErrorAfterCallCount int
		expectedErrorCode              string
	}{
		{
			name:                           "duplicate entry error",
			dbErrorAfterInsertCallCount:    0,
			dbError:                        &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			writeErrorAfterCallCount:       1,
			resolvePathErrorAfterCallCount: 1,
			expectedErrorCode:              ErrorCodeDuplicatedID,
		},
		{
			name:                           "generic database error",
			dbErrorAfterInsertCallCount:    0,
			dbError:                        errors.New("generic error"),
			writeErrorAfterCallCount:       1,
			resolvePathErrorAfterCallCount: 1,
			expectedErrorCode:              ErrorCodeDatabaseError,
		},
		{
			name:                           "storage write error",
			dbErrorAfterInsertCallCount:    1,
			writeErrorAfterCallCount:       0,
			resolvePathErrorAfterCallCount: 1,
			expectedErrorCode:              ErrorCodeStorageError,
		},
		{
			name:                           "resolve path error",
			dbErrorAfterInsertCallCount:    1,
			writeErrorAfterCallCount:       1,
			resolvePathErrorAfterCallCount: 0,
			expectedErrorCode:              ErrorCodeStorageError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payloadStorage := &payloadStorageMock{
				resolvePathErrorAfterCallCount: tc.resolvePathErrorAfterCallCount,
				writeErrorAfterCallCount:       tc.writeErrorAfterCallCount,
			}
			database := &databaseMock{
				errorAfterInsertCallCount: tc.dbErrorAfterInsertCallCount,
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

func TestService_Save_DeleteCalledOnError(t *testing.T) {
	t.Parallel()

	emailRequests := []EmailRequest{
		{
			MessageId:    "msg1",
			PayloadBytes: []byte("test payload"),
		},
	}

	testCases := []struct {
		name                  string
		dbErrorAfter          int
		dbError               error
		postCallbackError     error
		writeErrorAfter       int
		resolvePathErrorAfter int
		expectedDeleteCount   int
	}{
		{
			name:                  "delete called on commit error after write",
			dbErrorAfter:          1,
			postCallbackError:     errors.New("commit failed"),
			writeErrorAfter:       1,
			resolvePathErrorAfter: 1,
			expectedDeleteCount:   1,
		},
		{
			name:                  "delete not called on duplicate error",
			dbErrorAfter:          0,
			dbError:               &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			writeErrorAfter:       1,
			resolvePathErrorAfter: 1,
			expectedDeleteCount:   0,
		},
		{
			name:                  "delete not called on resolve path error",
			dbErrorAfter:          1,
			writeErrorAfter:       1,
			resolvePathErrorAfter: 0,
			expectedDeleteCount:   0,
		},
		{
			name:                  "delete not called on storage write error",
			dbErrorAfter:          1,
			writeErrorAfter:       0,
			resolvePathErrorAfter: 1,
			expectedDeleteCount:   0,
		},
		{
			name:                  "delete not called on generic database error",
			dbErrorAfter:          0,
			dbError:               errors.New("generic db error"),
			writeErrorAfter:       1,
			resolvePathErrorAfter: 1,
			expectedDeleteCount:   0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			payloadStorage := &payloadStorageMock{
				resolvePathErrorAfterCallCount: tc.resolvePathErrorAfter,
				writeErrorAfterCallCount:       tc.writeErrorAfter,
			}
			database := &databaseMock{
				errorAfterInsertCallCount: tc.dbErrorAfter,
				insertError:               tc.dbError,
				postCallbackError:         tc.postCallbackError,
			}

			sut := &Service{payloadStorage: payloadStorage, db: database}
			sut.Save(context.TODO(), emailRequests)

			assert.Equal(t, tc.expectedDeleteCount, payloadStorage.deleteCallCount)
		})
	}
}
