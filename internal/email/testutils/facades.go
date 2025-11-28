package testutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	testEmailTableStatusIndex = "StatusIndex"
	testEmailTableStatusMeta  = "_META"
	testEmailTableName        = "Outbox"
)

func NewAwsTestConfigFromEnv() aws.Config {
	return aws.Config{
		Region: os.Getenv("AWS_REGION"),
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		),
		BaseEndpoint: aws.String(os.Getenv("AWS_BASE_ENDPOINT")),
	}
}

type EmailTestFixtureKeys struct {
	Id     string
	Status string
}

type EmailTestFixture struct {
	Id     string
	Status string
	Latest string
}

type dynamodbRecord struct {
	Id         string                 `dynamodbav:"Id"`
	Status     string                 `dynamodbav:"Status"`
	Attributes map[string]interface{} `dynamodbav:"Attributes"`
	TTL        int64                  `dynamodbav:"TTL"`
}

type EmailDatabaseFacade struct {
	db *dynamodb.Client
}

func NewEmailDatabaseFacade() *EmailDatabaseFacade {
	cfg := NewAwsTestConfigFromEnv()
	return &EmailDatabaseFacade{db: dynamodb.NewFromConfig(cfg)}
}

func (edf *EmailDatabaseFacade) unmarshalList(src []map[string]types.AttributeValue) ([]EmailTestFixture, error) {
	var items []dynamodbRecord
	err := attributevalue.UnmarshalListOfMaps(src, &items)
	if err != nil {
		return []EmailTestFixture{}, err
	}

	emails := make([]EmailTestFixture, len(items))

	for i, item := range items {
		emails[i] = EmailTestFixture{
			Id:     item.Id,
			Status: item.Status,
			Latest: fmt.Sprint(item.Attributes["Latest"]),
		}
	}

	return emails, nil
}

func (edf *EmailDatabaseFacade) Query(ctx context.Context, status string, limit int) ([]EmailTestFixture, error) {
	query := fmt.Sprintf("SELECT Id, Status, Attributes, TTL FROM \"%v\".\"%v\" WHERE Status=? AND Attributes.Latest =?", testEmailTableName, testEmailTableStatusIndex)
	params, err := attributevalue.MarshalList([]interface{}{testEmailTableStatusMeta, status})
	if err != nil {
		return []EmailTestFixture{}, err
	}

	stmt := &dynamodb.ExecuteStatementInput{
		Parameters: params,
		Statement:  aws.String(query),
		Limit:      aws.Int32(int32(limit)),
	}

	res, err := edf.db.ExecuteStatement(ctx, stmt)
	if err != nil {
		return []EmailTestFixture{}, err
	}

	return edf.unmarshalList(res.Items)
}

func (edf *EmailDatabaseFacade) GetRecord(ctx context.Context, id string, status string) (dynamodbRecord, error) {
	query := fmt.Sprintf("SELECT Id, Status, Attributes, TTL FROM \"%v\" WHERE Id=? AND Status=?", testEmailTableName)
	params, err := attributevalue.MarshalList([]interface{}{id, status})
	if err != nil {
		return dynamodbRecord{}, err
	}

	stmt := &dynamodb.ExecuteStatementInput{
		Parameters: params,
		Statement:  aws.String(query),
	}

	res, err := edf.db.ExecuteStatement(ctx, stmt)
	if err != nil {
		return dynamodbRecord{}, err
	}

	if len(res.Items) == 0 {
		return dynamodbRecord{}, fmt.Errorf("record not found")
	}

	var record dynamodbRecord
	err = attributevalue.UnmarshalMap(res.Items[0], &record)
	if err != nil {
		return dynamodbRecord{}, err
	}

	return record, nil
}

func (edf *EmailDatabaseFacade) RemoveFixtures(fixtures []EmailTestFixtureKeys) error {
	if len(fixtures) == 0 {
		log.Print("no fixtures to delete")
		return nil
	}

	log.Printf("deleting fixtures: %v", fixtures)

	var errorList []error

	query := fmt.Sprintf("DELETE FROM \"%v\" WHERE Id=? AND Status=?", testEmailTableName)
	for _, fixture := range fixtures {
		params, _ := attributevalue.MarshalList([]interface{}{fixture.Id, fixture.Status})
		stmt := &dynamodb.ExecuteStatementInput{Statement: aws.String(query), Parameters: params}

		if _, err := edf.db.ExecuteStatement(context.TODO(), stmt); err != nil {
			wrappedErr := fmt.Errorf("error while deleting fixture %s, error: %w", fixture.Id, err)
			log.Printf("%v", wrappedErr)
			errorList = append(errorList, wrappedErr)
		}
	}

	if len(errorList) == 0 {
		return nil
	}

	var err error
	for _, foundError := range errorList {
		err = fmt.Errorf("%w\n%w", err, foundError)
	}

	return err
}

func (edf *EmailDatabaseFacade) InsertEmailWithStatus(ctx context.Context, id string, latestStatus string, updatedAtTime time.Time) error {
	ttl := time.Now().Add(30 * 24 * time.Hour).Unix()
	createdAt := time.Now().Format(time.RFC3339)
	updatedAt := updatedAtTime.Format(time.RFC3339)

	// Insert _META record
	metaStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?, 'TTL': ?}", testEmailTableName)
	metaAttrs := map[string]interface{}{
		"Latest":          latestStatus,
		"CreatedAt":       createdAt,
		"UpdatedAt":       updatedAt,
		"PayloadFilePath": "/test/path/payload.json",
	}
	metaParams, err := attributevalue.MarshalList([]interface{}{id, testEmailTableStatusMeta, metaAttrs, ttl})
	if err != nil {
		return err
	}

	// Insert status record
	statusStmt := fmt.Sprintf("INSERT INTO \"%v\" VALUE {'Id': ?, 'Status': ?, 'Attributes': ?, 'TTL': ?}", testEmailTableName)
	statusAttrs := map[string]interface{}{}
	statusParams, err := attributevalue.MarshalList([]interface{}{id, latestStatus, statusAttrs, ttl})
	if err != nil {
		return err
	}

	ti := &dynamodb.ExecuteTransactionInput{
		TransactStatements: []types.ParameterizedStatement{
			{Statement: aws.String(metaStmt), Parameters: metaParams},
			{Statement: aws.String(statusStmt), Parameters: statusParams},
		},
	}

	_, err = edf.db.ExecuteTransaction(ctx, ti)
	return err
}

