package email

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"multicarrier-email-api/internal/email/testutils"
)

var fixtures map[string]string

func deleteFixtures(t *testing.T, db *dynamodb.Client) {
	if len(fixtures) == 0 {
		t.Log("no fixtures to delete")
		return
	}

	t.Logf("deleting fixtures: %v", fixtures)

	query := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", "Outbox")
	for id, status := range fixtures {
		params, _ := attributevalue.MarshalList([]interface{}{id, status})
		stmt := &dynamodb.ExecuteStatementInput{Statement: aws.String(query), Parameters: params}

		if _, err := db.ExecuteStatement(context.TODO(), stmt); err != nil {
			t.Errorf("error while deleting fixture %s, error: %v", id, err)
		}
	}
}

func TestOutboxComponentWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	awsConfig := testutils.NewAwsTestConfigFromEnv()
	dynamo := dynamodb.NewFromConfig(awsConfig)
	sut := NewDatabase(dynamo, "Outbox", 30)

	fixtures = map[string]string{}
	defer deleteFixtures(t, dynamo)

	of := testutils.NewEmailDatabaseFacade()

	// no record in dynamo
	res, err := of.Query(context.TODO(), "ACCEPTED", 25)
	require.NoError(t, err)
	require.Len(t, res, 0)

	// insert two records in dynamo
	firstId := uuid.NewString()
	err = sut.Insert(context.TODO(), firstId, "/")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", firstId, err)
	fixtures[firstId] = "ACCEPTED"

	secondId := uuid.NewString()
	err = sut.Insert(context.TODO(), secondId, "/")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", secondId, err)
	fixtures[secondId] = "ACCEPTED"

	// should find 2 records with status pending
	res, err = of.Query(context.TODO(), "ACCEPTED", 25)
	require.NoError(t, err)
	require.Len(t, res, 2)

	// should not be able to insert again same id
	err = sut.Insert(context.TODO(), firstId, "/")
	require.Errorf(t, err, "inserted id %s, but it should have not because it's duplicated", firstId)
}

func TestReadyRecordHasTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	awsConfig := testutils.NewAwsTestConfigFromEnv()
	dynamo := dynamodb.NewFromConfig(awsConfig)
	sut := NewDatabase(dynamo, "Outbox", 30)

	fixtures = map[string]string{}
	defer deleteFixtures(t, dynamo)

	of := testutils.NewEmailDatabaseFacade()

	// insert a record
	testId := uuid.NewString()
	err := sut.Insert(context.TODO(), testId, "/")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", testId, err)
	fixtures[testId] = "ACCEPTED"

	// get the ACCEPTED record directly
	readyRecord, err := of.GetRecord(context.TODO(), testId, "ACCEPTED")
	require.NoError(t, err)
	require.Equal(t, testId, readyRecord.Id)
	require.Equal(t, "ACCEPTED", readyRecord.Status)

	// verify TTL is present in root
	require.Greater(t, readyRecord.TTL, int64(0))
	ttlValue := float64(readyRecord.TTL)
	require.Greater(t, ttlValue, float64(0), "TTL should be greater than 0")

	// verify TTL is in the future (at least current time)
	currentTime := time.Now().Unix()
	require.Greater(t, int64(ttlValue), currentTime, "TTL should be in the future")
}

func TestRequeueEmailForRequeuableStatuses(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	awsConfig := testutils.NewAwsTestConfigFromEnv()
	dynamo := dynamodb.NewFromConfig(awsConfig)
	sut := NewDatabase(dynamo, "Outbox", 30)

	fixtures = map[string]string{}
	defer deleteFixtures(t, dynamo)

	of := testutils.NewEmailDatabaseFacade()
	ctx := context.TODO()

	// Define stale time (more than 30 minutes ago)
	staleTime := time.Now().Add(-35 * time.Minute)

	// Test cases: states that can be requeued and their expected new status
	testCases := map[string]string{
		StatusIntaking:              StatusAccepted,
		StatusProcessing:            StatusReady,
		StatusCallingSentCallback:   StatusSent,
		StatusCallingFailedCallback: StatusFailed,
	}

	// Insert records with stale statuses
	insertedIds := make(map[string]string) // maps id -> currentStatus
	for currentStatus := range testCases {
		id := uuid.NewString()
		err := of.InsertEmailWithStatus(ctx, id, currentStatus, staleTime)
		require.NoErrorf(t, err, "failed inserting email with status %s", currentStatus)
		fixtures[id] = StatusMeta
		fixtures[id+currentStatus] = currentStatus
		insertedIds[id] = currentStatus
	}

	// Verify GetStaleEmails returns all 4 records
	staleEmails, err := sut.GetStaleEmails(ctx)
	require.NoError(t, err)
	require.Len(t, staleEmails, 4, "should find 4 stale emails")

	// Requeue each email and verify the status changes
	for _, staleEmail := range staleEmails {
		id := staleEmail.Id
		currentStatus := insertedIds[id]
		expectedNewStatus := testCases[currentStatus]

		// Requeue the email
		err := sut.RequeueEmail(ctx, id)
		require.NoErrorf(t, err, "failed to requeue email with status %s", currentStatus)

		// Verify the old status record is deleted
		_, err = of.GetRecord(ctx, id, currentStatus)
		require.Error(t, err, "old status record should be deleted")
		delete(fixtures, id+currentStatus)

		// Verify the _META record has been updated with new Latest
		metaRecord, err := of.GetRecord(ctx, id, StatusMeta)
		require.NoError(t, err)
		require.Equal(t, expectedNewStatus, metaRecord.Attributes["Latest"], "Latest should be updated to %s for status %s", expectedNewStatus, currentStatus)

		// Verify UpdatedAt has been updated
		updatedAt, ok := metaRecord.Attributes["UpdatedAt"].(string)
		require.True(t, ok)
		updatedAtParsed, err := time.Parse(time.RFC3339, updatedAt)
		require.NoError(t, err)
		require.True(t, updatedAtParsed.After(staleTime), "UpdatedAt should be more recent than the stale time")
	}

	// Verify GetStaleEmails now returns 0 records
	staleEmails, err = sut.GetStaleEmails(ctx)
	require.NoError(t, err)
	require.Len(t, staleEmails, 0, "should find 0 stale emails after requeue")
}

func TestRequeueEmailErrorsForNonRequeuableStatuses(t *testing.T) {
	if testing.Short() {
		t.Skip("component tests are skipped in short mode")
	}

	awsConfig := testutils.NewAwsTestConfigFromEnv()
	dynamo := dynamodb.NewFromConfig(awsConfig)
	sut := NewDatabase(dynamo, "Outbox", 30)

	fixtures = map[string]string{}
	defer deleteFixtures(t, dynamo)

	of := testutils.NewEmailDatabaseFacade()
	ctx := context.TODO()

	// Define recent time (less than 30 minutes ago, so they won't be in stale emails)
	recentTime := time.Now().Add(-5 * time.Minute)

	// States that should NOT be requeuable
	nonRequeuableStatuses := []string{
		StatusAccepted,
		StatusReady,
		StatusSent,
		StatusFailed,
		StatusInvalid,
		StatusSentAcknowledged,
		StatusFailedAcknowledged,
	}

	// Insert one stale email to verify GetStaleEmails still works
	staleId := uuid.NewString()
	staleTime := time.Now().Add(-35 * time.Minute)
	err := of.InsertEmailWithStatus(ctx, staleId, StatusIntaking, staleTime)
	require.NoError(t, err)
	fixtures[staleId] = StatusMeta
	fixtures[staleId+StatusIntaking] = StatusIntaking

	// Verify GetStaleEmails returns 1 record before testing
	staleEmails, err := sut.GetStaleEmails(ctx)
	require.NoError(t, err)
	initialStaleCount := len(staleEmails)
	require.Equal(t, 1, initialStaleCount, "should have 1 stale email initially")

	// Insert records with non-requeuable statuses
	for _, status := range nonRequeuableStatuses {
		id := uuid.NewString()
		err := of.InsertEmailWithStatus(ctx, id, status, recentTime)
		require.NoErrorf(t, err, "failed inserting email with status %s", status)
		fixtures[id] = StatusMeta
		fixtures[id+status] = status

		// Attempt to requeue should return error
		err = sut.RequeueEmail(ctx, id)
		require.Errorf(t, err, "requeue should fail for status %s", status)
		require.Contains(t, err.Error(), "cannot requeue email with Latest status", "error message should indicate invalid status")

		// Verify the status record still exists (not deleted)
		statusRecord, err := of.GetRecord(ctx, id, status)
		require.NoErrorf(t, err, "status record should still exist for %s", status)
		require.Equal(t, status, statusRecord.Status)

		// Verify GetStaleEmails still returns the same count
		staleEmails, err = sut.GetStaleEmails(ctx)
		require.NoError(t, err)
		require.Equal(t, initialStaleCount, len(staleEmails), "stale emails count should remain unchanged")
	}
}
