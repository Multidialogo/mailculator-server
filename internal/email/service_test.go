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

func TestService_Save(t *testing.T) {
	t.Parallel()

	emlBatch := testutils.DummyEMLDataBatch(2)

	type caseStruct struct {
		name                                   string
		emlStorageErrorAfterCallCount          int
		databaseErrorAfterInsertCallCount      int
		expectEMLStorageStoreCallCount         int
		expectedDatabaseInsertCallCount        int
		expectedDatabaseDeletePendingCallCount int
		expectError                            bool
	}

	testCases := []caseStruct{
		{
			name:                                   "success",
			emlStorageErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount:      2,
			expectEMLStorageStoreCallCount:         2,
			expectedDatabaseInsertCallCount:        2,
			expectedDatabaseDeletePendingCallCount: 0,
			expectError:                            false,
		},
		{
			name:                                   "storage error",
			emlStorageErrorAfterCallCount:          1,
			databaseErrorAfterInsertCallCount:      2,
			expectEMLStorageStoreCallCount:         2,
			expectedDatabaseInsertCallCount:        1,
			expectedDatabaseDeletePendingCallCount: 1,
			expectError:                            true,
		},
		{
			name:                                   "database error",
			emlStorageErrorAfterCallCount:          2,
			databaseErrorAfterInsertCallCount:      1,
			expectEMLStorageStoreCallCount:         2,
			expectedDatabaseInsertCallCount:        2,
			expectedDatabaseDeletePendingCallCount: 1,
			expectError:                            true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			emlStorage := &emlStorageMock{errorAfterCallCount: tc.emlStorageErrorAfterCallCount}
			database := &databaseMock{errorAfterInsertCallCount: tc.databaseErrorAfterInsertCallCount}

			sut := &Service{emlStorage: emlStorage, db: database}

			err := sut.Save(context.TODO(), emlBatch)
			assert.Equal(t, tc.expectEMLStorageStoreCallCount, emlStorage.callCount)
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
