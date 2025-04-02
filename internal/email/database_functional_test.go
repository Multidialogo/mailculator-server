package email

import (
	"context"
	"fmt"
	"testing"

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
	sut := New(dynamo)

	fixtures = map[string]string{}
	defer deleteFixtures(t, dynamo)

	of := testutils.NewEmailDatabaseFacade()

	// no record in dynamo
	res, err := of.Query(context.TODO(), "PENDING", 25)
	require.NoError(t, err)
	require.Len(t, res, 0)

	// insert two records in dynamo
	firstId := uuid.NewString()
	err = sut.Insert(context.TODO(), firstId, "/")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", firstId, err)
	fixtures[firstId] = "PENDING"

	secondId := uuid.NewString()
	err = sut.Insert(context.TODO(), secondId, "/")
	require.NoErrorf(t, err, "failed inserting id %s, error: %v", secondId, err)
	fixtures[secondId] = "PENDING"

	// should find 2 records with status pending
	res, err = of.Query(context.TODO(), "PENDING", 25)
	require.NoError(t, err)
	require.Len(t, res, 2)

	// should not be able to insert again same id
	err = sut.Insert(context.TODO(), firstId, "/")
	require.Errorf(t, err, "inserted id %s, but it should have not because it's duplicated", firstId)

	// delete pending
	err = sut.DeletePending(context.TODO(), firstId)
	require.NoError(t, err)
	delete(fixtures, firstId)

	// now it should be 1 record with status PENDING
	res, err = of.Query(context.TODO(), "PENDING", 25)
	require.NoError(t, err)
	require.Len(t, res, 1)

	// delete other pending
	err = sut.DeletePending(context.TODO(), secondId)
	require.NoError(t, err)
	delete(fixtures, secondId)

	// now it should be 0 record with status PENDING
	res, err = of.Query(context.TODO(), "PENDING", 25)
	require.NoError(t, err)
	require.Len(t, res, 0)
}
